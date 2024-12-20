package ws

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type LongConn interface {
	// Close this connection
	Close() error
	// WriteMessage Write message to connection,messageType means data type,can be set binary(2) and text(1).
	WriteMessage(messageType int, message []byte) error
	// ReadMessage Read message from connection.
	ReadMessage() (int, []byte, error)
	// GenerateLongConn Check the connection of the current and when it was sent are the same
	GenerateLongConn(w http.ResponseWriter, r *http.Request) error

	SetReadDeadline(timeout time.Duration) error
	SetWriteDeadline(timeout time.Duration) error
}

type GWebSocket struct {
	protocolType     int
	conn             *websocket.Conn
	handshakeTimeout time.Duration
	writeBufferSize  int
}

func newGWebSocket(protocolType int, handshakeTimeout time.Duration, wbs int) *GWebSocket {
	return &GWebSocket{protocolType: protocolType, handshakeTimeout: handshakeTimeout, writeBufferSize: wbs}
}

func (d *GWebSocket) GenerateLongConn(w http.ResponseWriter, r *http.Request) error {
	upgrader := &websocket.Upgrader{
		HandshakeTimeout: d.handshakeTimeout,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	if d.writeBufferSize > 0 { // default is 4kb.
		upgrader.WriteBufferSize = d.writeBufferSize
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// The upgrader.Upgrade method usually returns enough error messages to diagnose problems that may occur during the upgrade
		return errors.New("GenerateLongConn: WebSocket upgrade failed")
	}
	d.conn = conn
	return nil
}

func (d *GWebSocket) WriteMessage(messageType int, message []byte) error {
	// d.setSendConn(d.conn)
	return d.conn.WriteMessage(messageType, message)
}

func (d *GWebSocket) ReadMessage() (int, []byte, error) {
	return d.conn.ReadMessage()
}

func (d *GWebSocket) Close() error {
	return d.conn.Close()
}

func (d *GWebSocket) SetReadDeadline(timeout time.Duration) error {
	return d.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (d *GWebSocket) SetWriteDeadline(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	err := d.conn.SetWriteDeadline(time.Now().Add(timeout))
	return err
}

func (d *GWebSocket) RespondWithSuccess() error {
	if err := d.WriteMessage(MessageText, []byte("WS CONN Success")); err != nil {
		_ = d.Close()
		return errors.New("WriteMessage failed")
	}
	return nil
}
