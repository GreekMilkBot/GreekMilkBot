package greekmilkbot

import (
	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type Bot struct {
	Config  *Config
	Drivers []driver.Driver
}

func NewBot(config *Config) *Bot {
	return &Bot{
		Config:  config,
		Drivers: make([]driver.Driver, 0),
	}
}

func (b *Bot) AddDriver(driver driver.Driver) error {
	b.Drivers = append(b.Drivers, driver)
	return nil
}

func (b *Bot) Run() error {
	for _, driver := range b.Drivers {
		go driver.Run()
	}
	return nil
}
