package driver

import (
	"context"
	"time"
)

type DriverType int

const (
	DriverTypeHTTP DriverType = iota
	DriverTypeHTTPPost
	DriverTypeWebSocket
	DriverTypeWebSocketReverse
)

type Handler func(driver Driver, msg []byte)

type Driver interface {
	GetDriverType() DriverType
	Connect(ctx context.Context, handler Handler) error
	Send(msg string) error
}

type BaseDriver struct {
	DriverType DriverType
	Url        string
	Token      string

	Ttl time.Duration

	ReceiveChan chan string
	QuitChan    chan struct{}
}

func NewBaseDriver(driverType DriverType, url, token string) *BaseDriver {
	return &BaseDriver{
		DriverType:  driverType,
		Url:         url,
		Token:       token,
		Ttl:         15 * time.Second, // OneBot v11 默认心跳值
		ReceiveChan: make(chan string),
		QuitChan:    make(chan struct{}),
	}
}

func (d *BaseDriver) GetDriverType() DriverType {
	return d.DriverType
}

type DriverPacket struct {
	ID   string
	Data string
}
