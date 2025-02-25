package driver

import "context"

type Driver interface {
	Connect(ctx context.Context) error
	Send(msg string) error
	SetReceiveHandler(handler func(msg string))
}

type BaseDriver struct {
	Host string

	ReceiveChan    chan string
	QuitChan       chan struct{}
	ReceiveHandler func(string)
}

func NewBaseDriver(host string) *BaseDriver {
	return &BaseDriver{
		Host:        host,
		ReceiveChan: make(chan string),
		QuitChan:    make(chan struct{}),
	}
}

func (d *BaseDriver) SetReceiveHandler(handler func(msg string)) {
	d.ReceiveHandler = handler
}
