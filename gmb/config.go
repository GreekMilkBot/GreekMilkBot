package gmb

import (
	"github.com/GreekMilkBot/GreekMilkBot/bot"
)

type Config struct {
	Adapters []bot.Adapter
	Cache    int
}

func NewConfig(adapters ...bot.Adapter) *Config {
	return &Config{
		Adapters: adapters,
		Cache:    100,
	}
}
