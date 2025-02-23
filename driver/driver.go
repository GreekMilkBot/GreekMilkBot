package driver

import (
	"github.com/GreekMilkBot/GreekMilkBot/adapter"
)

type Driver interface {
	Run() error
	AddAdapter(adapter adapter.Adapter) error
}

type BaseDriver struct {
	Adapters []adapter.Adapter
}

func (d *BaseDriver) AddAdapter(adapter adapter.Adapter) error {
	d.Adapters = append(d.Adapters, adapter)
	return nil
}

func (d *BaseDriver) Run() error {
	return nil
}
