package tests

import (
	"context"
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
	ctx := context.Background()

	wsDriver := websocket.NewWebSocketDriver(os.Getenv("ONE_BOT_HOST"), os.Getenv("ONE_BOT_TOKEN"))
	testBot := gmb.NewGreekMilkBot(&gmb.Config{
		Adapters: []adapter.Adapter{onebotv11.NewOneBotV11Adapter(wsDriver)},
	})

	err := testBot.Run(ctx)
	assert.NoError(t, err)
	time.Sleep(30 * time.Second)
}
