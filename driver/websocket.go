package driver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// TODO: 支持断线重连
type WebSocketDriver struct {
	Url   string
	Token string
	Ttl   time.Duration

	conn  *atomic.Pointer[websocket.Conn]
	retry bool
	ctx   context.Context
}

func NewWebSocketDriver(context context.Context, url string, token string, retry bool) *WebSocketDriver {
	return &WebSocketDriver{
		ctx:   context,
		Url:   url,
		Token: token,
		Ttl:   30 * time.Second,
		retry: retry,
		conn:  new(atomic.Pointer[websocket.Conn]),
	}
}

func (d *WebSocketDriver) Bind(handler Handler) error {
	if d.conn.Load() != nil {
		return errors.New("WebSocketDriver already connected")
	}
	dialer := websocket.Dialer{}
	header := make(http.Header)
	if d.Token != "" {
		header.Add("Authorization", "Bearer "+d.Token)
	}
	conn, _, err := dialer.DialContext(d.ctx, d.Url, header)
	if err != nil {
		return err
	}
	if !d.conn.CompareAndSwap(nil, conn) {
		conn.Close()
		return errors.New("WebSocketDriver already connected")
	}
	go d.receive(conn, handler)
	return nil
}

func (d *WebSocketDriver) receive(conn *websocket.Conn, handler Handler) {
	defer conn.Close()
	defer d.conn.Store(nil)
l:
	for {
		select {
		case <-d.ctx.Done():
			break l
		default:
		}
		if d.Ttl > 0 {
			// set read timeout, avoid blocking
			_ = conn.SetReadDeadline(time.Now().Add(d.Ttl))
		}
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocketDriver: ReadMessage error", err)
			return
		}
		if messageType == websocket.TextMessage {
			handler(message)
		}
	}
}

func (d *WebSocketDriver) Send(msg string) error {
	conn := d.conn.Load()
	if conn == nil {
		return errors.New("WebSocket connection is not established")
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
