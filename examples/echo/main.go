package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	toolsMsg "github.com/GreekMilkBot/GreekMilkBot/tools/message"

	gmb "github.com/GreekMilkBot/GreekMilkBot"
	v11 "github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
	_ "github.com/GreekMilkBot/GreekMilkBot/tests/common"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	testBot, err := gmb.NewGreekMilkBot(
		gmb.WithAdapterURL(ctx, os.Getenv("TEST_BOT_URL")))
	if err != nil {
		panic(err)
	}
	testBot.HandleMessageFunc(func(ctx gmb.BotContext, message models.Message) {
		marshal, _ := json.MarshalIndent(&message, "", "  ")
		log.Infof(string(marshal))
		if strings.HasPrefix(message.Content.String(), "echo ") {
			contents := message.Content
			for i, content := range contents {
				if it, ok := content.(models.ContentText); ok {
					it.Text = strings.TrimPrefix(it.Text, "echo ")
					contents[i] = it
					break
				}
			}
			for i, content := range contents {
				if item, ok := content.(models.ContentUnknown); ok {
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
			clientMessage := toolsMsg.SenderMessage{
				QuoteID: "",
				Message: &contents,
			}
			if message.Quote != nil {
				clientMessage.QuoteID = message.Quote.ID
			}
			sender := toolsMsg.SenderClient(ctx)
			sendMessage, err := sender.SendMessage(&message, &clientMessage)
			if err != nil {
				log.Errorf("send message error %v", zap.Error(err))
			}
			log.Infof("新消息ID %s", sendMessage)
		}
	})
	testBot.HandleEventFunc(func(ctx gmb.BotContext, event models.Event) {
		content, _ := json.Marshal(event.Data)
		log.Infof("receive event[%v]: %s", event.Type, content)
	})
	if err := testBot.Run(ctx); err != nil {
		panic(err)
	}
}
