package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	gmb "github.com/GreekMilkBot/GreekMilkBot"
	onebotv11 "github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11"
	"github.com/GreekMilkBot/GreekMilkBot/driver/websocket"
)

func TestBot(t *testing.T) {
	ctx := context.Background()
	bot := gmb.NewBot(&gmb.Config{})
	wsDriver := websocket.NewWebSocketDriver("ws://10.0.0.200:4081")
	adapter := onebotv11.NewOneBotV11Adapter(wsDriver)
	bot.AddAdapter(adapter)
	err := bot.Run(ctx)
	assert.NoError(t, err)
}
