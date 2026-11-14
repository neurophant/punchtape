// Package relay implements the relay mode of the Gouija two-stage proxy.
// It acts as an HTTP proxy that forwards HTTPS CONNECT requests through
// an encrypted channel to the proxy node.
package relay

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/neurophant/gouija/internal/crypto"
	"github.com/neurophant/gouija/internal/logging"
	"github.com/neurophant/gouija/internal/pool"
	"github.com/neurophant/gouija/internal/throttle"
	"github.com/neurophant/gouija/internal/transport"
)

// Config holds relay configuration
type Config struct {
	ListenAddr string
	ProxyAddr  string
	Key        string
	Salt       string // Custom salt for key derivation (optional, uses default if empty)
	Transport  transport.Config
	PoolSize   int
	Timeout    time.Duration

	// HTTP Proxy Authentication credentials (optional)
	AuthUser string
	AuthPass string

	// Proxy node authentication (optional)
	RequireAuth bool // Whether to perform challenge-response authentication with proxy node

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

	// Bandwidth throttling (0 = unlimited)
	MaxBandwidth int64 // Maximum bandwidth in bytes per second
	BurstSize    int64 // Maximum burst size in bytes
}

// Server is the relay HTTP proxy server
type Server struct {
	config     Config
	pool       *pool.Pool
	httpServer *http.Server
	cipher     *crypto.Cipher
	authHeader string // Pre-computed expected auth header
	limiter    *throttle.Limiter
}

// New creates a new relay server
func New(cfg Config) (*Server, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 5
	}

	// Create cipher from key
	cipher, err := crypto.NewCipher(cfg.Key, cfg.Salt)
	if err != nil {
		return nil, fmt.Errorf("relay: failed to create cipher: %w", err)
	}

	// Create connection pool with DoS prevention settings
	p, err := pool.New(pool.Config{
		ProxyAddr:   cfg.ProxyAddr,
		Cipher:      cipher,
		Transport:   cfg.Transport,
		Size:        cfg.PoolSize,
		Timeout:     cfg.Timeout,
		RequireAuth: cfg.RequireAuth,
		Key:         cfg.Key,
		Salt:        cfg.Salt,

		// DoS prevention - Connection pacing
		ConnectionPacing:   cfg.ConnectionPacing,
		MaxConcurrentDials: cfg.MaxConcurrentDials,

		// DoS prevention - Exponential backoff
		BackoffInitial:    cfg.BackoffInitial,
		BackoffMax:        cfg.BackoffMax,
		BackoffMultiplier: cfg.BackoffMultiplier,
		BackoffJitter:     cfg.BackoffJitter,

		// Adaptive pool sizing
		MinPoolSize:     cfg.MinPoolSize,
		MaxPoolSize:     cfg.MaxPoolSize,
		GrowthThreshold: cfg.GrowthThreshold,
		ShrinkThreshold: cfg.ShrinkThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("relay: failed to create pool: %w", err)
	}

	// Create bandwidth limiter
	var limiter *throttle.Limiter
	if cfg.MaxBandwidth > 0 {
		limiter = throttle.NewSymmetricLimiter(cfg.MaxBandwidth)
		logging.Info("relay: bandwidth throttling enabled: %d bytes/sec", cfg.MaxBandwidth)
	}

	s := &Server{
		config:  cfg,
		pool:    p,
		cipher:  cipher,
		limiter: limiter,
	}

	// Pre-compute expected auth header if auth is enabled
	if cfg.AuthUser != "" && cfg.AuthPass != "" {
		credentials := base64.StdEncoding.EncodeToString([]byte(cfg.AuthUser + ":" + cfg.AuthPass))
		s.authHeader = "Basic " + credentials
	}

	s.httpServer = &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      http.HandlerFunc(s.handleRequest),
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	return s, nil
}

// Start starts the relay server
func (s *Server) Start() error {
	logging.Info("relay: starting server on %s", s.config.ListenAddr)
	logging.Info("relay: forwarding to proxy at %s (transport: %s)", s.config.ProxyAddr, s.config.Transport.Mode)
	logging.Info("relay: pool size: %d (min: %d, max: %d)", s.config.PoolSize,
		getOrDefault(s.config.MinPoolSize, pool.DefaultMinPoolSize),
		getOrDefault(s.config.MaxPoolSize, pool.DefaultMaxPoolSize))
	if s.authHeader != "" {
		logging.Info("relay: HTTP proxy authentication enabled")
	} else {
		logging.Info("relay: WARNING - HTTP proxy authentication disabled, anyone can use this proxy")
	}
	if s.config.RequireAuth {
		logging.Info("relay: proxy node authentication enabled (challenge-response)")
	}

	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("relay: server error: %w", err)
	}
	return nil
}

// getOrDefault returns the value or the default if zero
func getOrDefault(val, def int) int {
	if val <= 0 {
		return def
	}
	return val
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logging.Info("relay: shutting down server")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logging.Error("relay: HTTP server shutdown error: %v", err)
	}

	// Close pool
	if err := s.pool.Close(); err != nil {
		logging.Error("relay: pool close error: %v", err)
	}

	return nil
}

// handleRequest handles incoming HTTP requests
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Check authentication if enabled
	if s.authHeader != "" {
		if !s.checkAuth(r) {
			w.Header().Set("Proxy-Authenticate", `Basic realm="Gouija Proxy"`)
			http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
			logging.Info("relay: authentication failed from %s", r.RemoteAddr)
			return
		}
	}

	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
	} else {
		http.Error(w, "Only CONNECT method is supported", http.StatusMethodNotAllowed)
	}
}

// checkAuth verifies the Proxy-Authorization header
func (s *Server) checkAuth(r *http.Request) bool {
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	return constantTimeCompare(auth, s.authHeader)
}

// constantTimeCompare performs a constant-time string comparison
func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// handleConnect handles HTTP CONNECT requests for HTTPS tunneling
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	logging.Debug("relay: CONNECT request for %s from %s", r.Host, r.RemoteAddr)

	// Hijack the connection to get the raw TCP connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Server doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Create a tunnel through the pool
	tunnel := &tunnelHandler{
		clientConn: clientConn,
		host:       r.Host,
		pool:       s.pool,
		limiter:    s.limiter,
		done:       make(chan struct{}),
	}

	tunnel.run()
}

// tunnelHandler manages a single tunnel between client and proxy
type tunnelHandler struct {
	clientConn net.Conn
	host       string
	pool       *pool.Pool
	limiter    *throttle.Limiter
	tunnelID   uint32
	done       chan struct{}
	closeOnce  sync.Once
	mu         sync.Mutex
	connected  bool
	connectErr error
	connectCh  chan struct{}
}

// run establishes the tunnel and handles bidirectional data transfer
func (t *tunnelHandler) run() {
	defer t.clientConn.Close()

	t.connectCh = make(chan struct{})

	// Create tunnel through pool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tunnelID, err := t.pool.CreateTunnel(ctx, t.host, t.callback)
	if err != nil {
		logging.Debug("relay: failed to create tunnel to %s: %v", t.host, err)
		t.clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	t.tunnelID = tunnelID

	// Wait for connection result
	select {
	case <-t.connectCh:
		if t.connectErr != nil {
			logging.Debug("relay: tunnel %d connect failed: %v", t.tunnelID, t.connectErr)
			t.clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
			return
		}
	case <-time.After(30 * time.Second):
		logging.Debug("relay: tunnel %d connect timeout", t.tunnelID)
		t.clientConn.Write([]byte("HTTP/1.1 504 Gateway Timeout\r\n\r\n"))
		t.pool.CloseTunnel(t.tunnelID)
		return
	case <-t.done:
		return
	}

	// Send 200 OK to client
	_, err = t.clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		logging.Debug("relay: failed to send 200 OK: %v", err)
		t.pool.CloseTunnel(t.tunnelID)
		return
	}

	// Start reading from client and forwarding to tunnel
	t.readFromClient()
}

// callback handles data and errors from the pool tunnel
func (t *tunnelHandler) callback(tunnelID uint32, data []byte, err error) {
	if err != nil {
		t.mu.Lock()
		if !t.connected {
			// Connect error
			t.connectErr = err
			t.connected = true
			close(t.connectCh)
		}
		t.mu.Unlock()

		// Close the client connection
		t.close()
		return
	}

	if data == nil {
		// This is a successful connect response
		t.mu.Lock()
		if !t.connected {
			t.connected = true
			close(t.connectCh)
		}
		t.mu.Unlock()
		return
	}

	// Apply bandwidth throttling for downloads if enabled
	if t.limiter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := t.limiter.WaitDownload(ctx, len(data)); err != nil {
			cancel()
			logging.Debug("relay: download throttle timeout: %v", err)
			t.close()
			return
		}
		cancel()
	}

	// Write data to client
	if _, err := t.clientConn.Write(data); err != nil {
		logging.Debug("relay: failed to write to client: %v", err)
		t.close()
	}
}

// readFromClient reads data from the client and sends through the tunnel
func (t *tunnelHandler) readFromClient() {
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-t.done:
			return
		default:
		}

		n, err := t.clientConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				logging.Debug("relay: client read error: %v", err)
			}
			t.pool.CloseTunnel(t.tunnelID)
			t.close()
			return
		}

		if n > 0 {
			// Apply bandwidth throttling for uploads if enabled
			if t.limiter != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := t.limiter.WaitUpload(ctx, n); err != nil {
					cancel()
					logging.Debug("relay: upload throttle timeout: %v", err)
					t.pool.CloseTunnel(t.tunnelID)
					t.close()
					return
				}
				cancel()
			}

			if err := t.pool.SendData(t.tunnelID, buf[:n]); err != nil {
				logging.Debug("relay: failed to send data through tunnel: %v", err)
				t.close()
				return
			}
		}
	}
}

// close closes the tunnel
func (t *tunnelHandler) close() {
	t.closeOnce.Do(func() {
		close(t.done)
	})
}

// ParseProxyURL parses a proxy URL with optional credentials
// Format: http://user:pass@host:port or http://host:port
func ParseProxyURL(proxyURL string) (addr, user, pass string, err error) {
	// Remove scheme prefix
	url := strings.TrimPrefix(proxyURL, "http://")
	url = strings.TrimPrefix(url, "https://")

	// Check for credentials
	if atIdx := strings.LastIndex(url, "@"); atIdx != -1 {
		credentials := url[:atIdx]
		addr = url[atIdx+1:]

		if colonIdx := strings.Index(credentials, ":"); colonIdx != -1 {
			user = credentials[:colonIdx]
			pass = credentials[colonIdx+1:]
		} else {
			user = credentials
		}
	} else {
		addr = url
	}

	return addr, user, pass, nil
}
