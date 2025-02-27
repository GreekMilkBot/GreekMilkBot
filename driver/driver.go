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

type Driver interface {
	GetDriverType() DriverType
	Connect(ctx context.Context) error
	Send(msg string) error
	SetReceiveHandler(handler func(driver Driver, msg []byte))
}

type BaseDriver struct {
	DriverType DriverType
	Url        string
	Token      string

	Ttl time.Duration

	ReceiveChan    chan string
	QuitChan       chan struct{}
	ReceiveHandler func(Driver, []byte)
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

func (d *BaseDriver) SetReceiveHandler(handler func(driver Driver, msg []byte)) {
	d.ReceiveHandler = handler
}
