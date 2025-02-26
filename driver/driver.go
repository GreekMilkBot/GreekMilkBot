package driver

import "context"

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
	Host       string

	ReceiveChan    chan string
	QuitChan       chan struct{}
	ReceiveHandler func(Driver, []byte)
}

func NewBaseDriver(driverType DriverType, host string) *BaseDriver {
	return &BaseDriver{
		DriverType:  driverType,
		Host:        host,
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
