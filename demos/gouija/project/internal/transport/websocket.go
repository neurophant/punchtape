package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// wsConn wraps a WebSocket connection to implement the transport.Conn interface
type wsConn struct {
	conn       *websocket.Conn
	reader     *wsReader
	localAddr  net.Addr
	remoteAddr net.Addr
	mu         sync.Mutex
}

// wsReader handles reading from WebSocket as a continuous byte stream
type wsReader struct {
	conn   *websocket.Conn
	buf    []byte
	offset int
}

func (r *wsReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.buf) {
		// Need to read new message
		_, msg, err := r.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		r.buf = msg
		r.offset = 0
	}

	n := copy(p, r.buf[r.offset:])
	r.offset += n
	return n, nil
}

// newWSConn creates a new WebSocket transport connection
func newWSConn(conn *websocket.Conn) *wsConn {
	return &wsConn{
		conn:       conn,
		reader:     &wsReader{conn: conn},
		localAddr:  conn.LocalAddr(),
		remoteAddr: conn.RemoteAddr(),
	}
}

// Read reads data from the WebSocket connection
func (c *wsConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Write writes data to the WebSocket connection as a binary message
func (c *wsConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close closes the WebSocket connection
func (c *wsConn) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address
func (c *wsConn) LocalAddr() net.Addr {
	return c.localAddr
}

// RemoteAddr returns the remote network address
func (c *wsConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline sets the read and write deadlines
func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline
func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// wsDialer implements the Dialer interface for WebSocket connections
type wsDialer struct {
	path string
	tls  bool
}

// Dial establishes a WebSocket connection
func (d *wsDialer) Dial(address string) (Conn, error) {
	return d.DialTimeout(address, 0)
}

// DialTimeout establishes a WebSocket connection with timeout
func (d *wsDialer) DialTimeout(address string, timeout time.Duration) (Conn, error) {
	scheme := "ws"
	if d.tls {
		scheme = "wss"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   address,
		Path:   d.path,
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
	}

	if d.tls {
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	return newWSConn(conn), nil
}

// wsListener implements the Listener interface for WebSocket connections
type wsListener struct {
	ln       net.Listener
	server   *http.Server
	connCh   chan *websocket.Conn
	errCh    chan error
	upgrader websocket.Upgrader
	path     string
	closed   bool
	mu       sync.Mutex
}

// newWSListener creates a new WebSocket listener
func newWSListener(address string, cfg Config) (Listener, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	l := &wsListener{
		ln:     ln,
		connCh: make(chan *websocket.Conn, 100),
		errCh:  make(chan error, 1),
		path:   cfg.Path,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Accept all origins for proxy use
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Path, l.handleWebSocket)

	l.server = &http.Server{
		Handler: mux,
	}

	// Start HTTP server in background
	go func() {
		var err error
		if cfg.TLS && cfg.TLSCert != "" && cfg.TLSKey != "" {
			err = l.server.ServeTLS(ln, cfg.TLSCert, cfg.TLSKey)
		} else {
			err = l.server.Serve(ln)
		}
		if err != nil && err != http.ErrServerClosed {
			l.errCh <- err
		}
	}()

	return l, nil
}

// handleWebSocket handles incoming WebSocket connections
func (l *wsListener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	l.mu.Lock()
	closed := l.closed
	l.mu.Unlock()

	if closed {
		conn.Close()
		return
	}

	l.connCh <- conn
}

// Accept waits for and returns the next WebSocket connection
func (l *wsListener) Accept() (Conn, error) {
	select {
	case conn, ok := <-l.connCh:
		if !ok || conn == nil {
			return nil, net.ErrClosed
		}
		return newWSConn(conn), nil
	case err := <-l.errCh:
		return nil, err
	}
}

// Close closes the WebSocket listener
func (l *wsListener) Close() error {
	l.mu.Lock()
	l.closed = true
	l.mu.Unlock()

	close(l.connCh)
	return l.server.Close()
}

// Addr returns the listener's network address
func (l *wsListener) Addr() net.Addr {
	return l.ln.Addr()
}
