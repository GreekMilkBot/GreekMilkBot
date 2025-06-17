package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	_ "github.com/GreekMilkBot/GreekMilkBot/tests/common"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	testBot, err := gmb.NewGreekMilkBot(
		gmb.WithAdapterURL(ctx, os.Getenv("TEST_BOT_URL")))
	if err != nil {
		panic(err)
	}
	testBot.HandleMessageFunc(func(ctx context.Context, id string, message bot.Message) {
		marshal, _ := json.MarshalIndent(&message, "", "  ")
		log.Infof(string(marshal))
		if strings.HasPrefix(message.Content.String(), "echo ") {
			contents := message.Content
			for i, content := range contents {
				if it, ok := content.(bot.ContentText); ok {
					it.Text = strings.TrimPrefix(it.Text, "echo ")
					contents[i] = it
					break
				}
			}
			sendMessage, err := gmb.NewClientBus(id, testBot.ClientCall).SendMessage(&message, &contents)
			if err != nil {
				log.Errorf("send message error %v", zap.Error(err))
			}
			log.Infof(sendMessage)
		}
	})
	testBot.HandleEventFunc(func(ctx context.Context, id string, event bot.Event) {
	})
	if err := testBot.Run(ctx); err != nil {
		panic(err)
	}
}
