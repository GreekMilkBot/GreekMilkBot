package adapter

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type Adapter interface {
	ID() string
	Run(ctx Bus) error
}

type BaseAdapter struct {
	Driver driver.Driver

	Inited bool
	Bot    *bot.Bot
}

type Bus struct {
	ID string

	context.Context
}

func (b Bus) SendMessage(message bot.Message) error {

}
