package v11

import (
	"context"
	"fmt"
	"log"

	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
)

type OneBotV11Adapter struct {
	adapter.BaseAdapter
}

func NewOneBotV11Adapter(driver driver.Driver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		BaseAdapter: adapter.BaseAdapter{
			Driver: driver,
		},
	}
}

func (a *OneBotV11Adapter) Run(ctx context.Context) error {
	err := a.Driver.Connect(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	a.Driver.SetReceiveHandler(a.handleMessage)
	return nil
}

func (a *OneBotV11Adapter) handleMessage(msg string) {
	fmt.Printf("Adapter: Received message: %s\n", msg)
	go func(m string) {
		processed := a.processMessage(m)
		if err := a.Driver.Send(processed); err != nil {
			fmt.Printf("Adapter: Error sending message: %v\n", err)
		}
	}(msg)
}

// processMessage 模拟消息处理逻辑，可根据实际需求修改
func (a *OneBotV11Adapter) processMessage(msg string) string {
	processed := fmt.Sprintf("%s - processed", msg)
	fmt.Printf("Adapter: Processed message: %s\n", processed)
	return processed
}
