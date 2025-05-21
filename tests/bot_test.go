package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	onebotv11 "github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11"
	"github.com/GreekMilkBot/GreekMilkBot/driver/websocket"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
)

func TestBot(t *testing.T) {
	TestSetup()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	wsDriver := websocket.NewWebSocketDriver(os.Getenv("ONE_BOT_URL"), os.Getenv("ONE_BOT_TOKEN"))
	testBot := gmb.NewGreekMilkBot(&gmb.Config{
		Adapters: []adapter.Adapter{onebotv11.NewOneBotV11Adapter(wsDriver)},
	})

	err := testBot.Run(ctx)
	assert.NoError(t, err)
	for msg := range testBot.Receive() {
		fmt.Println(msg)
		//todo:
	}
	select {
	case <-ctx.Done():
	}
}
