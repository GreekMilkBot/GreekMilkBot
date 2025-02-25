package adapter

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type Adapter interface {
	Run(ctx context.Context) error
}

type BaseAdapter struct {
	Driver driver.Driver
}
