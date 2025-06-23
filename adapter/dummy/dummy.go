package dummy

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"net"
	"net/http"
	"net/url"
	"os"
)

//go:embed default.json
var defCfg []byte

func init() {
	gmb.RegisterAdapter("dummy", func(ctx context.Context, url url.URL) (bot.Adapter, error) {
		if url.Scheme != "bind" {
			return nil, errors.New("unsupported scheme :" + url.Scheme)
		}
		listen, err := net.Listen("tcp", url.Host)
		if err != nil {
			return nil, err
		}
		tree := NewTree()
		include := url.Query().Get("include")
		if include != "" {
			var data []byte
			if include == "default" {
				log.Infof("当前使用默认测试数据")
				data = defCfg
			} else {
				data, err = os.ReadFile(include)
				if err != nil {
					return nil, err
				}
			}
			if err := json.Unmarshal(data, tree); err != nil {
				return nil, err
			}
		}

		log.Infof(tree.String())

		go func() {
			select {
			case <-ctx.Done():
				_ = listen.Close()
			}
		}()
		go func() {
			_ = http.Serve(listen, tree)
		}()
		return &DummyAdapter{
			tree,
		}, nil
	})
}

type DummyAdapter struct {
	tree *Tree
}

func (d *DummyAdapter) Bind(ctx *bot.Bus) error {

	return nil
}
