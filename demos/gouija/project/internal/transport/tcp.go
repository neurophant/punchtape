package transport

import (
	"net"
	"time"
)

// TCPConfig holds TCP-specific configuration
type TCPConfig struct {
	// KeepAlive enables TCP keepalive probes
	KeepAlive bool

	// KeepAliveInterval sets the interval between keepalive probes
	// Default: 30 seconds
	KeepAliveInterval time.Duration
}

// DefaultTCPConfig returns the default TCP configuration
func DefaultTCPConfig() TCPConfig {
	return TCPConfig{
		KeepAlive:         true,
		KeepAliveInterval: 30 * time.Second,
	}
}

// tcpConn wraps net.Conn to implement the transport.Conn interface
type tcpConn struct {
	net.Conn
}

// tcpDialer implements the Dialer interface for TCP connections
type tcpDialer struct {
	tcpConfig TCPConfig
}

// Dial establishes a TCP connection
func (d *tcpDialer) Dial(address string) (Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// Apply TCP settings
	if err := d.applyTCPSettings(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &tcpConn{Conn: conn}, nil
}

// DialTimeout establishes a TCP connection with timeout
func (d *tcpDialer) DialTimeout(address string, timeout time.Duration) (Conn, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}

	// Apply TCP settings
	if err := d.applyTCPSettings(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &tcpConn{Conn: conn}, nil
}

// applyTCPSettings applies TCP-specific settings to the connection
func (d *tcpDialer) applyTCPSettings(conn net.Conn) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil // Not a TCP connection, skip settings
	}

	// Set keepalive
	if err := tcpConn.SetKeepAlive(d.tcpConfig.KeepAlive); err != nil {
		return err
	}

	// Set keepalive interval if keepalive is enabled
	if d.tcpConfig.KeepAlive && d.tcpConfig.KeepAliveInterval > 0 {
		if err := tcpConn.SetKeepAlivePeriod(d.tcpConfig.KeepAliveInterval); err != nil {
			// SetKeepAlivePeriod may not be supported on all platforms
			// Log but don't fail
			return nil
		}
	}

	return nil
}

// tcpListener wraps net.Listener to implement the transport.Listener interface
type tcpListener struct {
	net.Listener
	tcpConfig TCPConfig
}

// newTCPListener creates a new TCP listener
func newTCPListener(address string) (Listener, error) {
	return newTCPListenerWithConfig(address, DefaultTCPConfig())
}

// newTCPListenerWithConfig creates a new TCP listener with custom config
func newTCPListenerWithConfig(address string, cfg TCPConfig) (Listener, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpListener{
		Listener:  ln,
		tcpConfig: cfg,
	}, nil
}

// Accept waits for and returns the next TCP connection
func (l *tcpListener) Accept() (Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Apply TCP settings to accepted connection
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(l.tcpConfig.KeepAlive)
		if l.tcpConfig.KeepAlive && l.tcpConfig.KeepAliveInterval > 0 {
			tcpConn.SetKeepAlivePeriod(l.tcpConfig.KeepAliveInterval)
		}
	}

	return &tcpConn{Conn: conn}, nil
}
