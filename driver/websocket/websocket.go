package websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type WebSocketDriver struct {
	*driver.BaseDriver

	conn *websocket.Conn
}

func NewWebSocketDriver(host, token string) *WebSocketDriver {
	return &WebSocketDriver{
		BaseDriver: driver.NewBaseDriver(driver.DriverTypeWebSocketReverse, host, token),
	}
}

func (d *WebSocketDriver) Connect(ctx context.Context) error {
	if d.conn != nil {
		return nil
	}

	var err error
	dialer := websocket.Dialer{}
	header := make(http.Header)
	if d.Token != "" {
		header.Add("Authorization", "Bearer "+d.Token)
	}
	d.conn, _, err = dialer.DialContext(ctx, d.Host, header)
	if err != nil {
		return err
	}

	go d.receive(ctx)

	return nil
}

func (d *WebSocketDriver) receive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket receiver stopped due to cancellation.")
			d.conn.Close()
			return
		default:
		}

		if d.Ttl > 0 {
			// set read timeout, avoid blocking
			_ = d.conn.SetReadDeadline(time.Now().Add(d.Ttl))
		}
		messageType, message, err := d.conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocketDriver: ReadMessage error", err)
			return
		}
		if messageType == websocket.TextMessage && d.ReceiveHandler != nil {
			d.ReceiveHandler(d, message)
		}
	}
}

func (d *WebSocketDriver) Send(msg string) error {
	if d.conn == nil {
		return fmt.Errorf("WebSocket connection is not established")
	}

	return d.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
