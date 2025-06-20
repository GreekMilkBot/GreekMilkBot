package dummy

import (
	"context"
	"errors"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"net"
	"net/http"
	"net/url"
)

func init() {
	gmb.RegisterAdapter("dummy", func(ctx context.Context, url url.URL) (bot.Adapter, error) {
		if url.Scheme != "bind" {
			return nil, errors.New("unsupported scheme :" + url.Scheme)
		}
		listen, err := net.Listen("tcp", url.Host)
		if err != nil {
			return nil, err
		}
		mux := http.NewServeMux()
		go func() {
			select {
			case <-ctx.Done():
				_ = listen.Close()
			}
		}()
		go func() {
			_ = http.Serve(listen, mux)
		}()
		return &DummyAdapter{}, nil
	})
}

type DummyAdapter struct {
	tree *Tree
}

func (d *DummyAdapter) Bind(ctx *bot.Bus) error {

	return nil
}
