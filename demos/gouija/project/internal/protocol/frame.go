package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/neurophant/gouija/internal/crypto"
)

const (
	// FrameHeaderSize is the size of the length prefix (4 bytes)
	FrameHeaderSize = 4

	// MaxFrameSize is the maximum size of an encrypted frame (1MB)
	MaxFrameSize = 1024 * 1024
)

var (
	// ErrFrameTooLarge indicates a frame exceeds MaxFrameSize
	ErrFrameTooLarge = errors.New("protocol: frame too large")

	// ErrConnectionClosed indicates the connection was closed
	ErrConnectionClosed = errors.New("protocol: connection closed")
)

// bufferPool provides reusable buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 32*1024) // 32KB default buffer
		return &buf
	},
}

// Conn wraps a connection with encrypted message framing.
// It is safe for concurrent reads from one goroutine and writes from another,
// but not for concurrent reads or concurrent writes.
type Conn struct {
	rw     io.ReadWriteCloser
	cipher *crypto.Cipher

	// Read state
	readMu  sync.Mutex
	readBuf []byte

	// Write state
	writeMu  sync.Mutex
	writeBuf []byte
}

// NewConn wraps a connection with encrypted framing.
func NewConn(rw io.ReadWriteCloser, cipher *crypto.Cipher) *Conn {
	return &Conn{
		rw:       rw,
		cipher:   cipher,
		readBuf:  make([]byte, FrameHeaderSize),
		writeBuf: make([]byte, 0, 32*1024),
	}
}

// ReadMessage reads and decrypts a message from the connection.
func (c *Conn) ReadMessage() (*Message, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	// Read frame header (4-byte length)
	if _, err := io.ReadFull(c.rw, c.readBuf[:FrameHeaderSize]); err != nil {
		if err == io.EOF {
			return nil, ErrConnectionClosed
		}
		return nil, fmt.Errorf("protocol: failed to read frame header: %w", err)
	}

	frameLen := binary.BigEndian.Uint32(c.readBuf[:FrameHeaderSize])
	if frameLen > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}

	// Get a buffer from the pool or allocate if needed
	var frameBuf []byte
	if frameLen <= 32*1024 {
		bufPtr := bufferPool.Get().(*[]byte)
		frameBuf = (*bufPtr)[:frameLen]
		defer bufferPool.Put(bufPtr)
	} else {
		frameBuf = make([]byte, frameLen)
	}

	// Read encrypted frame
	if _, err := io.ReadFull(c.rw, frameBuf); err != nil {
		return nil, fmt.Errorf("protocol: failed to read frame: %w", err)
	}

	// Decrypt
	plaintext, err := c.cipher.Decrypt(frameBuf)
	if err != nil {
		return nil, fmt.Errorf("protocol: decryption failed: %w", err)
	}

	// Unmarshal message
	msg := &Message{}
	if err := msg.UnmarshalBinary(plaintext); err != nil {
		return nil, fmt.Errorf("protocol: failed to unmarshal message: %w", err)
	}

	return msg, nil
}

// WriteMessage encrypts and writes a message to the connection.
func (c *Conn) WriteMessage(msg *Message) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// Marshal message
	plaintext, err := msg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("protocol: failed to marshal message: %w", err)
	}

	// Encrypt
	ciphertext, err := c.cipher.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("protocol: encryption failed: %w", err)
	}

	frameLen := len(ciphertext)
	if frameLen > MaxFrameSize {
		return ErrFrameTooLarge
	}

	// Prepare write buffer with header + ciphertext
	c.writeBuf = c.writeBuf[:0]
	c.writeBuf = append(c.writeBuf, 0, 0, 0, 0) // placeholder for length
	binary.BigEndian.PutUint32(c.writeBuf[:FrameHeaderSize], uint32(frameLen))
	c.writeBuf = append(c.writeBuf, ciphertext...)

	// Write frame
	if _, err := c.rw.Write(c.writeBuf); err != nil {
		return fmt.Errorf("protocol: failed to write frame: %w", err)
	}

	return nil
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	return c.rw.Close()
}

// SendConnect sends a CONNECT message and returns the response.
func (c *Conn) SendConnect(tunnelID uint32, host string) error {
	msg, err := NewConnectMessage(tunnelID, host)
	if err != nil {
		return err
	}
	return c.WriteMessage(msg)
}

// SendConnectOK sends a CONNECT_OK message.
func (c *Conn) SendConnectOK(tunnelID uint32) error {
	return c.WriteMessage(NewConnectOKMessage(tunnelID))
}

// SendConnectErr sends a CONNECT_ERR message.
func (c *Conn) SendConnectErr(tunnelID uint32, errMsg string) error {
	return c.WriteMessage(NewConnectErrMessage(tunnelID, errMsg))
}

// SendData sends a DATA message.
func (c *Conn) SendData(tunnelID uint32, data []byte) error {
	return c.WriteMessage(NewDataMessage(tunnelID, data))
}

// SendClose sends a CLOSE message.
func (c *Conn) SendClose(tunnelID uint32) error {
	return c.WriteMessage(NewCloseMessage(tunnelID))
}

// SendPing sends a PING message.
func (c *Conn) SendPing() error {
	return c.WriteMessage(NewPingMessage())
}

// SendPong sends a PONG message.
func (c *Conn) SendPong() error {
	return c.WriteMessage(NewPongMessage())
}

// SendAuthChallenge sends an AUTH_CHALLENGE message with the given nonce.
func (c *Conn) SendAuthChallenge(challenge []byte) error {
	return c.WriteMessage(NewAuthChallengeMessage(challenge))
}

// SendAuthResponse sends an AUTH_RESPONSE message with the HMAC response.
func (c *Conn) SendAuthResponse(response []byte) error {
	return c.WriteMessage(NewAuthResponseMessage(response))
}

// SendAuthOK sends an AUTH_OK message.
func (c *Conn) SendAuthOK() error {
	return c.WriteMessage(NewAuthOKMessage())
}

// SendAuthFail sends an AUTH_FAIL message with the given reason.
func (c *Conn) SendAuthFail(reason string) error {
	return c.WriteMessage(NewAuthFailMessage(reason))
}
