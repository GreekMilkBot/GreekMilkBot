package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	v11 "github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11"

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
			for i, content := range contents {
				if item, ok := content.(bot.ContentUnknown); ok {
					// 处理 onebot 的自定义消息
					if strings.HasPrefix(item.Type, "onebot11_") {
						value := make(map[string]interface{})
						err := json.Unmarshal([]byte(item.Value), &value)
						if err != nil {
							continue
						}
						contents[i] = v11.OneBotCustomContent{
							Type: strings.TrimPrefix(item.Type, "onebot11_"),
							Data: value,
						}
					}
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
		content, _ := json.Marshal(event.Data)
		log.Infof("receive event[%v]: %s", event.Type, content)
	})
	if err := testBot.Run(ctx); err != nil {
		panic(err)
	}
}
