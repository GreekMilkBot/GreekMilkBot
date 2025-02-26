package adapter

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type Adapter interface {
	Run(ctx context.Context) error
}

type BaseAdapter struct {
	Driver driver.Driver

	Inited bool
	Bot    *bot.Bot
}
