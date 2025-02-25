package websocket

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type WebSocketDriver struct {
	*driver.BaseDriver

	conn *websocket.Conn
}

func NewWebSocketDriver(host string) *WebSocketDriver {
	return &WebSocketDriver{
		BaseDriver: driver.NewBaseDriver(host),
	}
}

func (d *WebSocketDriver) Connect(ctx context.Context) error {
	if d.conn != nil {
		return nil
	}

	var err error
	dialer := websocket.Dialer{}
	d.conn, _, err = dialer.DialContext(ctx, d.Host, nil)
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

		// 设置读取超时时间，避免长时间阻塞
		d.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		messageType, message, err := d.conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocketDriver: ReadMessage error", err)
			return
		}
		if messageType == websocket.TextMessage && d.ReceiveHandler != nil {
			d.ReceiveHandler(string(message))
		}
	}
}

func (d *WebSocketDriver) Send(msg string) error {
	if d.conn == nil {
		return fmt.Errorf("WebSocket connection is not established")
	}

	return d.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
