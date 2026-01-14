package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConn defines the interface for a WebSocket connection.
// This abstraction allows testing without real network connections.
type WebSocketConn interface {
	// ReadMessage reads a message from the connection.
	ReadMessage() (messageType int, p []byte, err error)
	// WriteMessage writes a message to the connection.
	WriteMessage(messageType int, data []byte) error
	// WriteControl writes a control message (ping, pong, close) to the connection.
	WriteControl(messageType int, data []byte, deadline time.Time) error
	// Close closes the connection.
	Close() error
	// SetReadDeadline sets the read deadline on the connection.
	SetReadDeadline(t time.Time) error
	// SetPingHandler sets the handler for ping messages received from the peer.
	SetPingHandler(h func(appData string) error)
}

// WebSocketDialer defines the interface for dialing WebSocket connections.
// This abstraction allows testing without real network connections.
type WebSocketDialer interface {
	// Dial creates a new WebSocket connection to the given URL.
	Dial(urlStr string, requestHeader http.Header) (WebSocketConn, *http.Response, error)
}

// GorillaWebSocketDialer is the default implementation using gorilla/websocket.
type GorillaWebSocketDialer struct {
	*websocket.Dialer
}

// NewGorillaDialer creates a new GorillaWebSocketDialer with settings optimized for
// networks that filter small TLS Client Hello packets.
// Go 1.24+ includes X25519MLKEM768 in default curve preferences, which creates a
// larger Client Hello (~1500 bytes) that passes through network filters.
// We explicitly set NextProtos to ["http/1.1"] to avoid HTTP/2 negotiation,
// as WebSocket requires HTTP/1.1.
func NewGorillaDialer() *GorillaWebSocketDialer {
	return &GorillaWebSocketDialer{
		Dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				// Force HTTP/1.1 - WebSocket requires HTTP/1.1, not HTTP/2
				// Without this, ALPN may negotiate h2, causing WebSocket handshake failure
				NextProtos: []string{"http/1.1"},
			},
		},
	}
}

// Dial implements WebSocketDialer using gorilla/websocket.
func (d *GorillaWebSocketDialer) Dial(urlStr string, requestHeader http.Header) (WebSocketConn, *http.Response, error) {
	conn, resp, err := d.Dialer.Dial(urlStr, requestHeader)
	if err != nil {
		return nil, resp, err
	}
	return conn, resp, nil
}

// EventSender is an interface for sending events back to the server.
type EventSender interface {
	SendEvent(msgType MessageType, data interface{}) error
}
