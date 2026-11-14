// Package protocol defines the wire protocol for communication
// between relay and proxy nodes in the Gouija two-stage proxy.
package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Message types for the wire protocol
const (
	MsgConnect    byte = 0x01 // Request to connect to target host:port
	MsgConnectOK  byte = 0x02 // Connection established successfully
	MsgConnectErr byte = 0x03 // Connection failed with error message
	MsgData       byte = 0x04 // Tunneled data payload
	MsgClose      byte = 0x05 // Close the tunnel
	MsgPing       byte = 0x06 // Keep-alive ping
	MsgPong       byte = 0x07 // Keep-alive pong response

	// Authentication message types
	MsgAuthChallenge byte = 0x10 // Server sends random challenge
	MsgAuthResponse  byte = 0x11 // Client sends HMAC response
	MsgAuthOK        byte = 0x12 // Authentication successful
	MsgAuthFail      byte = 0x13 // Authentication failed
)

// AuthChallengeSize is the size of the authentication challenge nonce
const AuthChallengeSize = 32

// MaxHostLen is the maximum length of a host:port string
const MaxHostLen = 255

// MaxPayloadSize is the maximum size of a message payload (64KB)
const MaxPayloadSize = 64 * 1024

var (
	// ErrInvalidMessage indicates a malformed message
	ErrInvalidMessage = errors.New("protocol: invalid message")

	// ErrHostTooLong indicates the host string exceeds MaxHostLen
	ErrHostTooLong = errors.New("protocol: host too long")

	// ErrPayloadTooLarge indicates the payload exceeds MaxPayloadSize
	ErrPayloadTooLarge = errors.New("protocol: payload too large")

	// ErrUnknownMessageType indicates an unknown message type
	ErrUnknownMessageType = errors.New("protocol: unknown message type")
)

// Message represents a protocol message
type Message struct {
	Type     byte
	TunnelID uint32
	Payload  []byte // For DATA: raw data; For CONNECT: host:port; For CONNECT_ERR: error message
}

// MarshalBinary serializes the message to binary format.
// Format: [Type:1][TunnelID:4][Payload:variable]
func (m *Message) MarshalBinary() ([]byte, error) {
	payloadLen := len(m.Payload)
	if payloadLen > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	// Allocate buffer: type(1) + tunnelID(4) + payload(n)
	buf := make([]byte, 1+4+payloadLen)

	buf[0] = m.Type
	binary.BigEndian.PutUint32(buf[1:5], m.TunnelID)
	copy(buf[5:], m.Payload)

	return buf, nil
}

// UnmarshalBinary deserializes a message from binary format.
func (m *Message) UnmarshalBinary(data []byte) error {
	if len(data) < 5 {
		return ErrInvalidMessage
	}

	m.Type = data[0]
	m.TunnelID = binary.BigEndian.Uint32(data[1:5])

	if len(data) > 5 {
		m.Payload = make([]byte, len(data)-5)
		copy(m.Payload, data[5:])
	} else {
		m.Payload = nil
	}

	return nil
}

// NewConnectMessage creates a CONNECT message for the given target host.
func NewConnectMessage(tunnelID uint32, host string) (*Message, error) {
	if len(host) > MaxHostLen {
		return nil, ErrHostTooLong
	}

	return &Message{
		Type:     MsgConnect,
		TunnelID: tunnelID,
		Payload:  []byte(host),
	}, nil
}

// NewConnectOKMessage creates a CONNECT_OK message.
func NewConnectOKMessage(tunnelID uint32) *Message {
	return &Message{
		Type:     MsgConnectOK,
		TunnelID: tunnelID,
	}
}

// NewConnectErrMessage creates a CONNECT_ERR message with an error description.
func NewConnectErrMessage(tunnelID uint32, errMsg string) *Message {
	return &Message{
		Type:     MsgConnectErr,
		TunnelID: tunnelID,
		Payload:  []byte(errMsg),
	}
}

// NewDataMessage creates a DATA message with the given payload.
func NewDataMessage(tunnelID uint32, data []byte) *Message {
	return &Message{
		Type:     MsgData,
		TunnelID: tunnelID,
		Payload:  data,
	}
}

// NewCloseMessage creates a CLOSE message.
func NewCloseMessage(tunnelID uint32) *Message {
	return &Message{
		Type:     MsgClose,
		TunnelID: tunnelID,
	}
}

// NewPingMessage creates a PING message.
func NewPingMessage() *Message {
	return &Message{
		Type: MsgPing,
	}
}

// NewPongMessage creates a PONG message.
func NewPongMessage() *Message {
	return &Message{
		Type: MsgPong,
	}
}

// NewAuthChallengeMessage creates an AUTH_CHALLENGE message with a random nonce.
func NewAuthChallengeMessage(challenge []byte) *Message {
	return &Message{
		Type:    MsgAuthChallenge,
		Payload: challenge,
	}
}

// NewAuthResponseMessage creates an AUTH_RESPONSE message with HMAC response.
func NewAuthResponseMessage(response []byte) *Message {
	return &Message{
		Type:    MsgAuthResponse,
		Payload: response,
	}
}

// NewAuthOKMessage creates an AUTH_OK message.
func NewAuthOKMessage() *Message {
	return &Message{
		Type: MsgAuthOK,
	}
}

// NewAuthFailMessage creates an AUTH_FAIL message with optional error.
func NewAuthFailMessage(reason string) *Message {
	return &Message{
		Type:    MsgAuthFail,
		Payload: []byte(reason),
	}
}

// GetChallenge extracts the challenge from an AUTH_CHALLENGE message.
func (m *Message) GetChallenge() []byte {
	if m.Type != MsgAuthChallenge {
		return nil
	}
	return m.Payload
}

// GetAuthResponse extracts the response from an AUTH_RESPONSE message.
func (m *Message) GetAuthResponse() []byte {
	if m.Type != MsgAuthResponse {
		return nil
	}
	return m.Payload
}

// GetAuthFailReason extracts the reason from an AUTH_FAIL message.
func (m *Message) GetAuthFailReason() string {
	if m.Type != MsgAuthFail {
		return ""
	}
	return string(m.Payload)
}

// GetHost extracts the host string from a CONNECT message payload.
func (m *Message) GetHost() string {
	if m.Type != MsgConnect {
		return ""
	}
	return string(m.Payload)
}

// GetError extracts the error message from a CONNECT_ERR message payload.
func (m *Message) GetError() string {
	if m.Type != MsgConnectErr {
		return ""
	}
	return string(m.Payload)
}

// TypeString returns a human-readable name for the message type.
func (m *Message) TypeString() string {
	switch m.Type {
	case MsgConnect:
		return "CONNECT"
	case MsgConnectOK:
		return "CONNECT_OK"
	case MsgConnectErr:
		return "CONNECT_ERR"
	case MsgData:
		return "DATA"
	case MsgClose:
		return "CLOSE"
	case MsgPing:
		return "PING"
	case MsgPong:
		return "PONG"
	case MsgAuthChallenge:
		return "AUTH_CHALLENGE"
	case MsgAuthResponse:
		return "AUTH_RESPONSE"
	case MsgAuthOK:
		return "AUTH_OK"
	case MsgAuthFail:
		return "AUTH_FAIL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", m.Type)
	}
}

// String returns a string representation of the message for debugging.
func (m *Message) String() string {
	switch m.Type {
	case MsgConnect:
		return fmt.Sprintf("%s(tunnel=%d, host=%s)", m.TypeString(), m.TunnelID, m.GetHost())
	case MsgConnectErr:
		return fmt.Sprintf("%s(tunnel=%d, err=%s)", m.TypeString(), m.TunnelID, m.GetError())
	case MsgData:
		return fmt.Sprintf("%s(tunnel=%d, len=%d)", m.TypeString(), m.TunnelID, len(m.Payload))
	default:
		return fmt.Sprintf("%s(tunnel=%d)", m.TypeString(), m.TunnelID)
	}
}
