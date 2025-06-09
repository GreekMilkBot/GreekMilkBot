package tests

import (
	"context"
	"encoding/json"
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
	testBot.HandleMessageFunc(func(ctx context.Context, id string, message bot.Message) {
		marshal, _ := json.MarshalIndent(&message, "", "  ")
		log.Info(string(marshal))
		if strings.HasPrefix(message.Content.String(), "echo ") {

			sendMessage, err := gmb.NewClientBus(id, testBot.ClientCall).SendMessage(&message, &message.Content)
			assert.NoError(t, err)
			log.Info(sendMessage)
		}
	})
	testBot.HandleEventFunc(func(ctx context.Context, id string, event bot.Event) {
	})
	go func() {
		err := testBot.Run(ctx)
		assert.NoError(t, err)
	}()
	select {
	case <-ctx.Done():
	}
}
