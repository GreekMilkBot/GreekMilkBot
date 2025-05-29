package adapter

import (
	"context"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type Adapter interface {
	Run(ctx Bus) error
}

type BaseAdapter struct {
	Driver driver.Driver

	Inited bool
	Bot    *bot.Bot
}

type Bus struct {
	ID string

	Tx chan<- bot.Packet
	Rx <-chan bot.Packet

	context.Context
}

func (b Bus) SendMessage(message bot.Message) error {
	b.Tx <- bot.Packet{
		Plugin: b.ID,
		Type:   bot.PacketMessage,
		Data:   message,
	}
	return nil
}
