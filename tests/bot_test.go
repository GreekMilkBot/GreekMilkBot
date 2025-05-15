package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"

	"github.com/stretchr/testify/assert"

	onebotv11 "github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11"
	"github.com/GreekMilkBot/GreekMilkBot/driver/websocket"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
)

func TestBot(t *testing.T) {
	TestSetup()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	wsDriver := websocket.NewWebSocketDriver(os.Getenv("ONE_BOT_URL"), os.Getenv("ONE_BOT_TOKEN"))
	testBot := gmb.NewGreekMilkBot(gmb.NewConfig(onebotv11.NewOneBotV11Adapter(wsDriver)))

	err := testBot.Run(ctx)
	assert.NoError(t, err)
	testBot.HandleMessageFunc(func(ctx context.Context, message bot.Message) {
		log.Info(message.Content.String())
		if strings.HasPrefix(message.Content.String(), "echo ") {
			sendMessage, err := testBot.WithBot(ctx).SendMessage(&message, message.Content)
			assert.NoError(t, err)
			log.Info(sendMessage)
		}
	})

	testBot.HandleEventFunc(func(ctx context.Context, message bot.Event) {
	})

	select {
	case <-ctx.Done():
	}
}
