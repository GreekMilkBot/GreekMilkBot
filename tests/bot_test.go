package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	onebotv11 "github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11"
	"github.com/GreekMilkBot/GreekMilkBot/driver/websocket"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

func TestBot(t *testing.T) {
	log.SetLevel(zap.DebugLevel)

	ctx := context.Background()

	wsDriver := websocket.NewWebSocketDriver("ws://10.0.0.200:4081")
	testBot := gmb.NewGreekMilkBot(&gmb.Config{
		Adapters: []adapter.Adapter{onebotv11.NewOneBotV11Adapter(wsDriver)},
	})

	err := testBot.Run(ctx)
	assert.NoError(t, err)
	time.Sleep(30 * time.Second)
}
