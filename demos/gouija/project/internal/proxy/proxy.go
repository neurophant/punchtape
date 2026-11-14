// Package proxy implements the proxy mode of the Gouija two-stage proxy.
// It accepts encrypted connections from relay nodes, decrypts traffic,
// and forwards requests to target HTTP/HTTPS resources.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/neurophant/gouija/internal/crypto"
	"github.com/neurophant/gouija/internal/logging"
	"github.com/neurophant/gouija/internal/protocol"
	"github.com/neurophant/gouija/internal/transport"
)

// Authentication errors
var (
	ErrAuthFailed   = errors.New("proxy: authentication failed")
	ErrAuthTimeout  = errors.New("proxy: authentication timeout")
	ErrAuthRequired = errors.New("proxy: authentication required")
)

// Config holds proxy configuration
type Config struct {
	ListenAddr  string
	Key         string
	Salt        string // Custom salt for key derivation (optional, uses default if empty)
	Transport   transport.Config
	Timeout     time.Duration
	RequireAuth bool // Whether to require challenge-response authentication
}

// Server is the proxy server that accepts relay connections
type Server struct {
	config   Config
	cipher   *crypto.Cipher
	authKey  []byte // Derived key for HMAC authentication
	listener transport.Listener
	wg       sync.WaitGroup
	closed   chan struct{}
}

// New creates a new proxy server
func New(cfg Config) (*Server, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":9000"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Create cipher from key
	cipher, err := crypto.NewCipher(cfg.Key, cfg.Salt)
	if err != nil {
		return nil, fmt.Errorf("proxy: failed to create cipher: %w", err)
	}

	// Derive auth key if authentication is required
	var authKey []byte
	if cfg.RequireAuth {
		authKey = crypto.DeriveAuthKey(cfg.Key, cfg.Salt)
	}

	return &Server{
		config:  cfg,
		cipher:  cipher,
		authKey: authKey,
		closed:  make(chan struct{}),
	}, nil
}

// Start starts the proxy server
func (s *Server) Start() error {
	var err error
	s.listener, err = transport.NewListener(s.config.Transport, s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("proxy: failed to create listener: %w", err)
	}

	authStatus := "disabled"
	if s.config.RequireAuth {
		authStatus = "enabled"
	}
	logging.Info("proxy: starting server on %s (transport: %s, auth: %s)", s.config.ListenAddr, s.config.Transport.Mode, authStatus)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.closed:
				return nil
			default:
				logging.Error("proxy: accept error: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logging.Info("proxy: shutting down server")
	close(s.closed)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all connections to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Info("proxy: all connections closed")
	case <-ctx.Done():
		logging.Info("proxy: shutdown timeout, forcing close")
	}

	return nil
}

// handleConnection handles a connection from a relay node
func (s *Server) handleConnection(rawConn io.ReadWriteCloser) {
	defer s.wg.Done()
	defer rawConn.Close()

	// Try to get remote addr for logging
	remoteAddr := "unknown"
	if rc, ok := rawConn.(interface{ RemoteAddr() net.Addr }); ok {
		remoteAddr = rc.RemoteAddr().String()
	}
	logging.Debug("proxy: new connection from %s", remoteAddr)

	// Wrap with protocol connection
	conn := protocol.NewConn(rawConn, s.cipher)

	// Perform authentication handshake if required
	if s.config.RequireAuth {
		if err := s.performAuthHandshake(conn); err != nil {
			logging.Info("proxy: authentication failed for %s: %v", remoteAddr, err)
			return
		}
		logging.Debug("proxy: authentication successful for %s", remoteAddr)
	}

	// Create connection handler
	handler := &connHandler{
		conn:    conn,
		tunnels: make(map[uint32]*tunnel),
		timeout: s.config.Timeout,
		closed:  s.closed,
	}

	handler.run()
}

// performAuthHandshake performs challenge-response authentication
func (s *Server) performAuthHandshake(conn *protocol.Conn) error {
	// Generate challenge
	challenge, err := crypto.GenerateChallenge()
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Send AUTH_CHALLENGE
	if err := conn.SendAuthChallenge(challenge); err != nil {
		return fmt.Errorf("failed to send challenge: %w", err)
	}

	// Wait for AUTH_RESPONSE
	msg, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if msg.Type != protocol.MsgAuthResponse {
		// Send AUTH_FAIL for unexpected message
		conn.SendAuthFail("expected AUTH_RESPONSE")
		return fmt.Errorf("expected AUTH_RESPONSE, got %d", msg.Type)
	}

	// Verify response
	response := msg.GetAuthResponse()
	if !crypto.VerifyAuthResponse(s.authKey, challenge, response) {
		conn.SendAuthFail("invalid response")
		return ErrAuthFailed
	}

	// Send AUTH_OK
	if err := conn.SendAuthOK(); err != nil {
		return fmt.Errorf("failed to send AUTH_OK: %w", err)
	}

	return nil
}

// connHandler handles messages on a single relay connection
type connHandler struct {
	conn    *protocol.Conn
	tunnels map[uint32]*tunnel
	mu      sync.RWMutex
	timeout time.Duration
	closed  chan struct{}
}

// tunnel represents an active tunnel to a target
type tunnel struct {
	id        uint32
	conn      net.Conn
	handler   *connHandler
	done      chan struct{}
	closeOnce sync.Once
}

// run processes messages from the relay connection
func (h *connHandler) run() {
	for {
		select {
		case <-h.closed:
			h.closeAllTunnels()
			return
		default:
		}

		msg, err := h.conn.ReadMessage()
		if err != nil {
			if err != protocol.ErrConnectionClosed {
				logging.Debug("proxy: read error: %v", err)
			}
			h.closeAllTunnels()
			return
		}

		h.handleMessage(msg)
	}
}

// handleMessage dispatches a message to the appropriate handler
func (h *connHandler) handleMessage(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MsgConnect:
		h.handleConnect(msg)

	case protocol.MsgData:
		h.handleData(msg)

	case protocol.MsgClose:
		h.handleClose(msg)

	case protocol.MsgPing:
		if err := h.conn.SendPong(); err != nil {
			logging.Debug("proxy: failed to send pong: %v", err)
		}

	case protocol.MsgPong:
		// Nothing to do

	default:
		logging.Debug("proxy: unknown message type: %d", msg.Type)
	}
}

// handleConnect handles a CONNECT message
func (h *connHandler) handleConnect(msg *protocol.Message) {
	host := msg.GetHost()
	tunnelID := msg.TunnelID

	logging.Debug("proxy: tunnel %d connecting to %s", tunnelID, host)

	// Connect to target
	targetConn, err := net.DialTimeout("tcp", host, h.timeout)
	if err != nil {
		logging.Debug("proxy: tunnel %d failed to connect to %s: %v", tunnelID, host, err)
		h.conn.SendConnectErr(tunnelID, err.Error())
		return
	}

	// Create tunnel
	t := &tunnel{
		id:      tunnelID,
		conn:    targetConn,
		handler: h,
		done:    make(chan struct{}),
	}

	h.mu.Lock()
	h.tunnels[tunnelID] = t
	h.mu.Unlock()

	// Send connect OK
	if err := h.conn.SendConnectOK(tunnelID); err != nil {
		logging.Debug("proxy: failed to send connect OK: %v", err)
		t.close()
		return
	}

	// Start reading from target
	go t.readFromTarget()
}

// handleData handles a DATA message
func (h *connHandler) handleData(msg *protocol.Message) {
	h.mu.RLock()
	t, ok := h.tunnels[msg.TunnelID]
	h.mu.RUnlock()

	if !ok {
		logging.Debug("proxy: received data for unknown tunnel %d", msg.TunnelID)
		return
	}

	select {
	case <-t.done:
		return
	default:
	}

	// Write data to target
	if _, err := t.conn.Write(msg.Payload); err != nil {
		logging.Debug("proxy: tunnel %d write error: %v", msg.TunnelID, err)
		t.close()
	}
}

// handleClose handles a CLOSE message
func (h *connHandler) handleClose(msg *protocol.Message) {
	h.mu.Lock()
	t, ok := h.tunnels[msg.TunnelID]
	if ok {
		delete(h.tunnels, msg.TunnelID)
	}
	h.mu.Unlock()

	if ok {
		logging.Debug("proxy: tunnel %d closed by relay", msg.TunnelID)
		t.close()
	}
}

// closeAllTunnels closes all active tunnels
func (h *connHandler) closeAllTunnels() {
	h.mu.Lock()
	tunnels := make([]*tunnel, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		tunnels = append(tunnels, t)
	}
	h.tunnels = make(map[uint32]*tunnel)
	h.mu.Unlock()

	for _, t := range tunnels {
		t.close()
	}
}

// readFromTarget reads data from the target and sends to relay
func (t *tunnel) readFromTarget() {
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-t.done:
			return
		default:
		}

		n, err := t.conn.Read(buf)
		if err != nil {
			// Check if tunnel was already closed by relay (expected close)
			select {
			case <-t.done:
				// Tunnel was closed by relay, this is expected - just return
				return
			default:
				// Tunnel wasn't closed yet, this is an actual error
				if err != io.EOF {
					logging.Debug("proxy: tunnel %d target read error: %v", t.id, err)
				}
				t.sendCloseAndRemove()
				return
			}
		}

		if n > 0 {
			if err := t.handler.conn.SendData(t.id, buf[:n]); err != nil {
				logging.Debug("proxy: tunnel %d failed to send data: %v", t.id, err)
				t.sendCloseAndRemove()
				return
			}
		}
	}
}

// sendCloseAndRemove sends a close message and removes the tunnel
func (t *tunnel) sendCloseAndRemove() {
	t.handler.mu.Lock()
	delete(t.handler.tunnels, t.id)
	t.handler.mu.Unlock()

	t.handler.conn.SendClose(t.id)
	t.close()
}

// close closes the tunnel
func (t *tunnel) close() {
	t.closeOnce.Do(func() {
		close(t.done)
		t.conn.Close()
	})
}
