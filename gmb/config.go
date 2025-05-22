package gmb

import "github.com/GreekMilkBot/GreekMilkBot/adapter"

type Config struct {
	Adapters []adapter.Adapter
	Cache    int
}

func NewConfig(adapters ...adapter.Adapter) *Config {
	return &Config{
		Adapters: adapters,
		Cache:    100,
	}
}
