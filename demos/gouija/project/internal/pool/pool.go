// Package pool provides a connection pool with multiplexing support
// for the Gouija two-stage proxy relay node.
package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/neurophant/gouija/internal/crypto"
	"github.com/neurophant/gouija/internal/logging"
	"github.com/neurophant/gouija/internal/protocol"
	"github.com/neurophant/gouija/internal/transport"
)

var (
	// ErrPoolClosed indicates the pool has been closed
	ErrPoolClosed = errors.New("pool: pool is closed")

	// ErrNoAvailableConnection indicates no connections are available
	ErrNoAvailableConnection = errors.New("pool: no available connection")

	// ErrTunnelClosed indicates the tunnel was closed
	ErrTunnelClosed = errors.New("pool: tunnel closed")

	// ErrConnectFailed indicates the remote connection failed
	ErrConnectFailed = errors.New("pool: connect failed")

	// ErrAuthFailed indicates authentication failed
	ErrAuthFailed = errors.New("pool: authentication failed")
)

// Default configuration values for DoS prevention
const (
	// DefaultConnectionPacing is the delay between establishing new connections
	DefaultConnectionPacing = 500 * time.Millisecond

	// DefaultMaxConcurrentDials limits simultaneous dial attempts
	DefaultMaxConcurrentDials = 2

	// DefaultBackoffInitial is the initial reconnection backoff interval
	DefaultBackoffInitial = 5 * time.Second

	// DefaultBackoffMax is the maximum reconnection backoff interval
	DefaultBackoffMax = 120 * time.Second

	// DefaultBackoffMultiplier is the backoff multiplier
	DefaultBackoffMultiplier = 2.0

	// DefaultBackoffJitter is the jitter factor (±20%)
	DefaultBackoffJitter = 0.2

	// DefaultMinPoolSize is the minimum number of connections
	DefaultMinPoolSize = 1

	// DefaultMaxPoolSize is the maximum number of connections
	DefaultMaxPoolSize = 10

	// DefaultGrowthThreshold triggers pool growth when utilization exceeds this
	DefaultGrowthThreshold = 0.8

	// DefaultShrinkThreshold triggers pool shrink when utilization below this
	DefaultShrinkThreshold = 0.2
)

// TunnelCallback is called when data arrives on a tunnel or the tunnel closes.
type TunnelCallback func(tunnelID uint32, data []byte, err error)

// tunnelInfo tracks an active tunnel
type tunnelInfo struct {
	callback TunnelCallback
	closed   atomic.Bool
}

// connBackoff tracks backoff state for a connection slot
type connBackoff struct {
	failures    int
	nextAttempt time.Time
	mu          sync.Mutex
}

// reset resets the backoff state after successful connection
func (b *connBackoff) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.nextAttempt = time.Time{}
}

// fail records a failure and calculates next attempt time
func (b *connBackoff) fail(cfg *backoffConfig) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++

	// Calculate backoff with exponential increase
	backoff := float64(cfg.Initial) * math.Pow(cfg.Multiplier, float64(b.failures-1))
	if backoff > float64(cfg.Max) {
		backoff = float64(cfg.Max)
	}

	// Add jitter (±jitter%)
	jitterRange := backoff * cfg.Jitter
	jitter := (rand.Float64()*2 - 1) * jitterRange
	backoff += jitter

	duration := time.Duration(backoff)
	b.nextAttempt = time.Now().Add(duration)

	return duration
}

// canAttempt returns true if enough time has passed for a retry
func (b *connBackoff) canAttempt() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return time.Now().After(b.nextAttempt)
}

// getFailures returns the current failure count
func (b *connBackoff) getFailures() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.failures
}

// backoffConfig holds backoff configuration
type backoffConfig struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     float64
}

// poolConn represents a single connection in the pool
type poolConn struct {
	conn         *protocol.Conn
	rawConn      io.ReadWriteCloser
	tunnels      sync.Map // tunnelID -> *tunnelInfo
	tunnelCount  atomic.Int32
	closed       atomic.Bool
	pool         *Pool
	createdAt    time.Time
	lastActivity atomic.Int64 // Unix timestamp of last activity
}

// updateActivity updates the last activity timestamp
func (pc *poolConn) updateActivity() {
	pc.lastActivity.Store(time.Now().Unix())
}

// Pool manages a pool of encrypted connections to the proxy node
type Pool struct {
	proxyAddr   string
	cipher      *crypto.Cipher
	dialer      transport.Dialer
	timeout     time.Duration
	requireAuth bool
	authKey     []byte

	// Pool sizing
	minSize int
	maxSize int
	size    int // Current target size

	// Connection pacing
	connectionPacing   time.Duration
	maxConcurrentDials int
	dialSemaphore      chan struct{}

	// Backoff configuration
	backoffCfg *backoffConfig

	// Adaptive sizing thresholds
	growthThreshold float64
	shrinkThreshold float64

	// Connection state
	conns        []*poolConn
	connBackoffs []*connBackoff
	connIdx      atomic.Uint32
	nextTunnelID atomic.Uint32 // Global tunnel ID counter to ensure uniqueness across connections
	mu           sync.RWMutex
	closed       atomic.Bool
	closeOnce    sync.Once
	wg           sync.WaitGroup

	// Metrics for adaptive sizing
	totalTunnels atomic.Int32
}

// Config holds pool configuration
type Config struct {
	ProxyAddr   string
	Cipher      *crypto.Cipher
	Transport   transport.Config
	Size        int
	Timeout     time.Duration
	RequireAuth bool   // Whether to perform authentication handshake
	Key         string // Pre-shared key for authentication (used to derive authKey)
	Salt        string // Custom salt for auth key derivation (optional)

	// DoS prevention - Connection pacing
	ConnectionPacing   time.Duration // Delay between new connections (default: 500ms)
	MaxConcurrentDials int           // Max simultaneous dial attempts (default: 2)

	// DoS prevention - Exponential backoff
	BackoffInitial    time.Duration // Initial backoff interval (default: 5s)
	BackoffMax        time.Duration // Maximum backoff interval (default: 120s)
	BackoffMultiplier float64       // Backoff multiplier (default: 2.0)
	BackoffJitter     float64       // Jitter factor ±percentage (default: 0.2)

	// Adaptive pool sizing
	MinPoolSize     int     // Minimum connections (default: 1)
	MaxPoolSize     int     // Maximum connections (default: 10)
	GrowthThreshold float64 // Grow when utilization exceeds (default: 0.8)
	ShrinkThreshold float64 // Shrink when utilization below (default: 0.2)
}

// applyDefaults fills in default values for unset configuration options
func (cfg *Config) applyDefaults() {
	if cfg.Size <= 0 {
		cfg.Size = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Connection pacing defaults
	if cfg.ConnectionPacing <= 0 {
		cfg.ConnectionPacing = DefaultConnectionPacing
	}
	if cfg.MaxConcurrentDials <= 0 {
		cfg.MaxConcurrentDials = DefaultMaxConcurrentDials
	}

	// Backoff defaults
	if cfg.BackoffInitial <= 0 {
		cfg.BackoffInitial = DefaultBackoffInitial
	}
	if cfg.BackoffMax <= 0 {
		cfg.BackoffMax = DefaultBackoffMax
	}
	if cfg.BackoffMultiplier <= 0 {
		cfg.BackoffMultiplier = DefaultBackoffMultiplier
	}
	if cfg.BackoffJitter <= 0 {
		cfg.BackoffJitter = DefaultBackoffJitter
	}

	// Adaptive sizing defaults
	if cfg.MinPoolSize <= 0 {
		cfg.MinPoolSize = DefaultMinPoolSize
	}
	if cfg.MaxPoolSize <= 0 {
		cfg.MaxPoolSize = DefaultMaxPoolSize
	}
	// Ensure Size is within bounds
	if cfg.Size < cfg.MinPoolSize {
		cfg.Size = cfg.MinPoolSize
	}
	if cfg.Size > cfg.MaxPoolSize {
		cfg.Size = cfg.MaxPoolSize
	}
	if cfg.GrowthThreshold <= 0 {
		cfg.GrowthThreshold = DefaultGrowthThreshold
	}
	if cfg.ShrinkThreshold <= 0 {
		cfg.ShrinkThreshold = DefaultShrinkThreshold
	}
}

// New creates a new connection pool with lazy connection initialization
func New(cfg Config) (*Pool, error) {
	cfg.applyDefaults()

	dialer := transport.NewDialer(cfg.Transport)

	// Derive auth key if authentication is required
	var authKey []byte
	if cfg.RequireAuth {
		authKey = crypto.DeriveAuthKey(cfg.Key, cfg.Salt)
	}

	p := &Pool{
		proxyAddr:          cfg.ProxyAddr,
		cipher:             cfg.Cipher,
		dialer:             dialer,
		timeout:            cfg.Timeout,
		requireAuth:        cfg.RequireAuth,
		authKey:            authKey,
		minSize:            cfg.MinPoolSize,
		maxSize:            cfg.MaxPoolSize,
		size:               cfg.Size,
		connectionPacing:   cfg.ConnectionPacing,
		maxConcurrentDials: cfg.MaxConcurrentDials,
		dialSemaphore:      make(chan struct{}, cfg.MaxConcurrentDials),
		backoffCfg: &backoffConfig{
			Initial:    cfg.BackoffInitial,
			Max:        cfg.BackoffMax,
			Multiplier: cfg.BackoffMultiplier,
			Jitter:     cfg.BackoffJitter,
		},
		growthThreshold: cfg.GrowthThreshold,
		shrinkThreshold: cfg.ShrinkThreshold,
		conns:           make([]*poolConn, cfg.MaxPoolSize),
		connBackoffs:    make([]*connBackoff, cfg.MaxPoolSize),
	}

	// Initialize backoff trackers for all slots
	for i := 0; i < cfg.MaxPoolSize; i++ {
		p.connBackoffs[i] = &connBackoff{}
	}

	logging.Info("pool: initializing with target size %d (min: %d, max: %d)", cfg.Size, cfg.MinPoolSize, cfg.MaxPoolSize)
	logging.Info("pool: connection pacing: %v, max concurrent dials: %d", cfg.ConnectionPacing, cfg.MaxConcurrentDials)
	logging.Info("pool: backoff: initial=%v, max=%v, multiplier=%.1f, jitter=%.0f%%",
		cfg.BackoffInitial, cfg.BackoffMax, cfg.BackoffMultiplier, cfg.BackoffJitter*100)

	// Start connection establishment with pacing
	p.wg.Add(1)
	go p.initConnectionsWithPacing()

	// Start background reconnection goroutine
	p.wg.Add(1)
	go p.reconnectLoop()

	// Start adaptive sizing goroutine
	p.wg.Add(1)
	go p.adaptiveSizingLoop()

	return p, nil
}

// initConnectionsWithPacing establishes initial connections with delays between them
func (p *Pool) initConnectionsWithPacing() {
	defer p.wg.Done()

	connectedCount := 0

	for i := 0; i < p.size; i++ {
		if p.closed.Load() {
			return
		}

		// Acquire dial semaphore to limit concurrent dials
		select {
		case p.dialSemaphore <- struct{}{}:
		default:
			// Wait for a slot
			p.dialSemaphore <- struct{}{}
		}

		pc, err := p.createConnection()
		<-p.dialSemaphore // Release semaphore

		if err != nil {
			logging.Debug("pool: connection %d failed (will retry with backoff): %v", i, err)
			p.connBackoffs[i].fail(p.backoffCfg)
			p.conns[i] = nil
		} else {
			p.mu.Lock()
			p.conns[i] = pc
			p.mu.Unlock()
			p.connBackoffs[i].reset()
			connectedCount++
			logging.Debug("pool: connection %d established", i)
		}

		// Add pacing delay between connection attempts (except after last)
		if i < p.size-1 && !p.closed.Load() {
			time.Sleep(p.connectionPacing)
		}
	}

	if connectedCount == 0 {
		logging.Info("pool: no connections established yet, will connect when proxy becomes available")
	} else {
		logging.Info("pool: %d/%d connections established", connectedCount, p.size)
	}
}

// reconnectLoop periodically tries to reconnect failed connections with exponential backoff
func (p *Pool) reconnectLoop() {
	defer p.wg.Done()

	// Use a shorter ticker for checking backoff timers
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if p.closed.Load() {
			return
		}
		p.tryReconnect()
	}
}

// tryReconnect attempts to reconnect connections that are ready for retry
func (p *Pool) tryReconnect() {
	p.mu.Lock()
	currentSize := p.size
	p.mu.Unlock()

	for i := 0; i < currentSize; i++ {
		if p.closed.Load() {
			return
		}

		p.mu.RLock()
		pc := p.conns[i]
		needsReconnect := pc == nil || pc.closed.Load()
		p.mu.RUnlock()

		if !needsReconnect {
			continue
		}

		// Check if backoff period has elapsed
		if !p.connBackoffs[i].canAttempt() {
			continue
		}

		// Try to acquire dial semaphore (non-blocking)
		select {
		case p.dialSemaphore <- struct{}{}:
			// Got semaphore, try to connect
			newPc, err := p.createConnection()
			<-p.dialSemaphore // Release semaphore

			if err != nil {
				backoff := p.connBackoffs[i].fail(p.backoffCfg)
				failures := p.connBackoffs[i].getFailures()
				logging.Debug("pool: connection %d reconnect failed (attempt %d, next retry in %v): %v",
					i, failures, backoff, err)
				continue
			}

			p.mu.Lock()
			p.conns[i] = newPc
			p.mu.Unlock()
			p.connBackoffs[i].reset()
			logging.Info("pool: connection %d reconnected", i)

			// Add pacing delay after successful reconnection
			time.Sleep(p.connectionPacing)

		default:
			// Semaphore full, skip this iteration
			continue
		}
	}
}

// adaptiveSizingLoop monitors pool utilization and adjusts size
func (p *Pool) adaptiveSizingLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if p.closed.Load() {
			return
		}
		p.adjustPoolSize()
	}
}

// adjustPoolSize grows or shrinks the pool based on utilization
func (p *Pool) adjustPoolSize() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate current utilization
	activeConns := 0
	totalTunnels := 0
	for i := 0; i < p.size; i++ {
		pc := p.conns[i]
		if pc != nil && !pc.closed.Load() {
			activeConns++
			totalTunnels += int(pc.tunnelCount.Load())
		}
	}

	if activeConns == 0 {
		return // No active connections, don't adjust
	}

	// Calculate tunnels per connection (utilization proxy)
	avgTunnelsPerConn := float64(totalTunnels) / float64(activeConns)
	utilization := avgTunnelsPerConn / 100.0 // Assume 100 tunnels per conn is "full"
	if utilization > 1.0 {
		utilization = 1.0
	}

	// Also consider connection availability
	connUtilization := float64(activeConns) / float64(p.size)

	// Use the higher of the two metrics
	if connUtilization > utilization {
		utilization = connUtilization
	}

	// Grow pool if utilization is high
	if utilization > p.growthThreshold && p.size < p.maxSize {
		newSize := p.size + 1
		logging.Info("pool: growing pool from %d to %d (utilization: %.1f%%)", p.size, newSize, utilization*100)
		p.size = newSize
	}

	// Shrink pool if utilization is low (but keep minimum)
	if utilization < p.shrinkThreshold && p.size > p.minSize {
		// Only shrink if we have idle connections
		if activeConns >= p.size-1 {
			newSize := p.size - 1
			logging.Info("pool: shrinking pool from %d to %d (utilization: %.1f%%)", p.size, newSize, utilization*100)
			p.size = newSize

			// Close excess connection if any
			if p.conns[newSize] != nil {
				p.conns[newSize].close()
				p.conns[newSize] = nil
			}
		}
	}
}

// createConnection creates a new pool connection
func (p *Pool) createConnection() (*poolConn, error) {
	rawConn, err := p.dialer.DialTimeout(p.proxyAddr, p.timeout)
	if err != nil {
		return nil, err
	}

	conn := protocol.NewConn(rawConn, p.cipher)

	// Perform authentication handshake if required
	if p.requireAuth {
		if err := p.performAuthHandshake(conn); err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	pc := &poolConn{
		conn:      conn,
		rawConn:   rawConn,
		pool:      p,
		createdAt: time.Now(),
	}
	pc.updateActivity()

	// Start reader goroutine
	p.wg.Add(1)
	go pc.readLoop()

	return pc, nil
}

// performAuthHandshake performs the client-side authentication handshake
func (p *Pool) performAuthHandshake(conn *protocol.Conn) error {
	// Wait for AUTH_CHALLENGE from server
	msg, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}

	if msg.Type != protocol.MsgAuthChallenge {
		return fmt.Errorf("expected AUTH_CHALLENGE, got %s", msg.TypeString())
	}

	challenge := msg.GetChallenge()
	if len(challenge) == 0 {
		return fmt.Errorf("empty challenge received")
	}

	// Compute HMAC response
	response := crypto.ComputeAuthResponse(p.authKey, challenge)

	// Send AUTH_RESPONSE
	if err := conn.SendAuthResponse(response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	// Wait for AUTH_OK or AUTH_FAIL
	msg, err = conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read auth result: %w", err)
	}

	switch msg.Type {
	case protocol.MsgAuthOK:
		return nil
	case protocol.MsgAuthFail:
		return fmt.Errorf("%w: %s", ErrAuthFailed, msg.GetAuthFailReason())
	default:
		return fmt.Errorf("unexpected message type: %s", msg.TypeString())
	}
}

// getConnection returns a connection from the pool using round-robin,
// creating new connections on demand if needed
func (p *Pool) getConnection() (*poolConn, error) {
	if p.closed.Load() {
		return nil, ErrPoolClosed
	}

	// First try with read lock (fast path)
	p.mu.RLock()
	currentSize := p.size
	idx := p.connIdx.Add(1) % uint32(currentSize)
	pc := p.conns[idx]

	if pc != nil && !pc.closed.Load() {
		p.mu.RUnlock()
		pc.updateActivity()
		return pc, nil
	}

	// Try to find any available connection
	for i := 0; i < currentSize; i++ {
		pc = p.conns[(int(idx)+i)%currentSize]
		if pc != nil && !pc.closed.Load() {
			p.mu.RUnlock()
			pc.updateActivity()
			return pc, nil
		}
	}
	p.mu.RUnlock()

	// No available connection, try to create one (slow path)
	return p.getOrCreateConnection()
}

// getOrCreateConnection tries to create a new connection for an empty slot
func (p *Pool) getOrCreateConnection() (*poolConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed.Load() {
		return nil, ErrPoolClosed
	}

	// Check again after acquiring write lock (another goroutine might have created one)
	for i := 0; i < p.size; i++ {
		pc := p.conns[i]
		if pc != nil && !pc.closed.Load() {
			pc.updateActivity()
			return pc, nil
		}
	}

	// Find an empty slot and try to create a connection
	// Use dial semaphore to prevent connection storms
	select {
	case p.dialSemaphore <- struct{}{}:
		defer func() { <-p.dialSemaphore }()

		for i := 0; i < p.size; i++ {
			if p.conns[i] == nil || p.conns[i].closed.Load() {
				// Check backoff
				if !p.connBackoffs[i].canAttempt() {
					continue
				}

				pc, err := p.createConnection()
				if err != nil {
					p.connBackoffs[i].fail(p.backoffCfg)
					return nil, fmt.Errorf("%w: %v", ErrNoAvailableConnection, err)
				}
				p.conns[i] = pc
				p.connBackoffs[i].reset()
				logging.Info("pool: connection %d established on demand", i)
				return pc, nil
			}
		}

	default:
		// Dial semaphore is full, another dial is in progress
		// Return error to let caller retry
	}

	return nil, ErrNoAvailableConnection
}

// CreateTunnel creates a new tunnel to the target host through the proxy.
// The callback is invoked when data arrives on the tunnel or when it closes.
func (p *Pool) CreateTunnel(ctx context.Context, host string, callback TunnelCallback) (uint32, error) {
	pc, err := p.getConnection()
	if err != nil {
		return 0, err
	}

	// Generate globally unique tunnel ID to prevent collisions across connections
	tunnelID := p.nextTunnelID.Add(1)

	// Register tunnel before sending CONNECT
	info := &tunnelInfo{callback: callback}
	pc.tunnels.Store(tunnelID, info)
	pc.tunnelCount.Add(1)
	p.totalTunnels.Add(1)

	// Send CONNECT message
	if err := pc.conn.SendConnect(tunnelID, host); err != nil {
		pc.tunnels.Delete(tunnelID)
		pc.tunnelCount.Add(-1)
		p.totalTunnels.Add(-1)
		return 0, fmt.Errorf("pool: failed to send connect: %w", err)
	}

	return tunnelID, nil
}

// SendData sends data through a tunnel
func (p *Pool) SendData(tunnelID uint32, data []byte) error {
	pc, info := p.findTunnel(tunnelID)
	if pc == nil || info == nil {
		return ErrTunnelClosed
	}

	if info.closed.Load() {
		return ErrTunnelClosed
	}

	pc.updateActivity()
	return pc.conn.SendData(tunnelID, data)
}

// CloseTunnel closes a tunnel
func (p *Pool) CloseTunnel(tunnelID uint32) error {
	pc, info := p.findTunnel(tunnelID)
	if pc == nil || info == nil {
		return nil // Already closed
	}

	info.closed.Store(true)
	pc.tunnels.Delete(tunnelID)
	pc.tunnelCount.Add(-1)
	p.totalTunnels.Add(-1)

	return pc.conn.SendClose(tunnelID)
}

// findTunnel finds the connection and info for a tunnel
func (p *Pool) findTunnel(tunnelID uint32) (*poolConn, *tunnelInfo) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for i := 0; i < p.size; i++ {
		pc := p.conns[i]
		if pc == nil {
			continue
		}
		if v, ok := pc.tunnels.Load(tunnelID); ok {
			return pc, v.(*tunnelInfo)
		}
	}
	return nil, nil
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		TargetSize:   p.size,
		MinSize:      p.minSize,
		MaxSize:      p.maxSize,
		TotalTunnels: int(p.totalTunnels.Load()),
	}

	for i := 0; i < p.size; i++ {
		pc := p.conns[i]
		if pc != nil && !pc.closed.Load() {
			stats.ActiveConns++
			stats.TunnelsPerConn = append(stats.TunnelsPerConn, int(pc.tunnelCount.Load()))
		}
	}

	return stats
}

// PoolStats contains pool statistics
type PoolStats struct {
	TargetSize     int
	MinSize        int
	MaxSize        int
	ActiveConns    int
	TotalTunnels   int
	TunnelsPerConn []int
}

// Close closes the pool and all connections
func (p *Pool) Close() error {
	p.closeOnce.Do(func() {
		p.closed.Store(true)

		p.mu.Lock()
		for i := 0; i < p.maxSize; i++ {
			if p.conns[i] != nil {
				p.conns[i].close()
			}
		}
		p.mu.Unlock()
	})

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logging.Info("pool: timeout waiting for goroutines to finish")
	}

	return nil
}

// readLoop reads messages from the connection and dispatches to tunnels
func (pc *poolConn) readLoop() {
	defer pc.pool.wg.Done()
	defer pc.close()

	for {
		msg, err := pc.conn.ReadMessage()
		if err != nil {
			if !pc.closed.Load() {
				logging.Debug("pool: read error: %v", err)
				pc.notifyAllTunnels(err)
			}
			return
		}

		pc.updateActivity()
		pc.handleMessage(msg)
	}
}

// handleMessage dispatches a message to the appropriate tunnel
func (pc *poolConn) handleMessage(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MsgConnectOK:
		// Connection established - notify callback with nil data to signal success
		logging.Debug("pool: tunnel %d connected", msg.TunnelID)
		if v, ok := pc.tunnels.Load(msg.TunnelID); ok {
			info := v.(*tunnelInfo)
			info.callback(msg.TunnelID, nil, nil)
		}

	case protocol.MsgConnectErr:
		// Connection failed
		if v, ok := pc.tunnels.Load(msg.TunnelID); ok {
			info := v.(*tunnelInfo)
			info.closed.Store(true)
			pc.tunnels.Delete(msg.TunnelID)
			pc.tunnelCount.Add(-1)
			pc.pool.totalTunnels.Add(-1)
			info.callback(msg.TunnelID, nil, fmt.Errorf("%w: %s", ErrConnectFailed, msg.GetError()))
		}

	case protocol.MsgData:
		// Data received
		if v, ok := pc.tunnels.Load(msg.TunnelID); ok {
			info := v.(*tunnelInfo)
			if !info.closed.Load() {
				info.callback(msg.TunnelID, msg.Payload, nil)
			}
		}

	case protocol.MsgClose:
		// Tunnel closed by remote
		if v, ok := pc.tunnels.Load(msg.TunnelID); ok {
			info := v.(*tunnelInfo)
			info.closed.Store(true)
			pc.tunnels.Delete(msg.TunnelID)
			pc.tunnelCount.Add(-1)
			pc.pool.totalTunnels.Add(-1)
			info.callback(msg.TunnelID, nil, ErrTunnelClosed)
		}

	case protocol.MsgPing:
		// Respond to ping
		if err := pc.conn.SendPong(); err != nil {
			logging.Debug("pool: failed to send pong: %v", err)
		}

	case protocol.MsgPong:
		// Pong received - nothing to do
	}
}

// notifyAllTunnels notifies all tunnels of an error
func (pc *poolConn) notifyAllTunnels(err error) {
	pc.tunnels.Range(func(key, value interface{}) bool {
		info := value.(*tunnelInfo)
		if !info.closed.Load() {
			info.closed.Store(true)
			pc.tunnelCount.Add(-1)
			pc.pool.totalTunnels.Add(-1)
			info.callback(key.(uint32), nil, err)
		}
		return true
	})
}

// close closes the pool connection
func (pc *poolConn) close() {
	if pc.closed.Swap(true) {
		return // Already closed
	}
	pc.conn.Close()
}
