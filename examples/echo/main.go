package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"

	_ "github.com/GreekMilkBot/GreekMilkBot/adapters"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/tools"

	gmb "github.com/GreekMilkBot/GreekMilkBot"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
	"go.uber.org/zap"
)

func main() {
	log.SetLevel(zapcore.DebugLevel)
	ctx := context.Background()
	pUrl := os.Getenv("BOT_URL")
	if pUrl == "" {
		pUrl = "dummy+bind://127.0.0.1:8080?include=default"
	}
	testBot, err := gmb.NewGreekMilkBot(
		gmb.WithPluginURL(ctx, pUrl))
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
			clientMessage := tools.SenderMessage{
				QuoteID: "",
				Message: &contents,
			}
			if message.Quote != nil {
				clientMessage.QuoteID = message.Quote.ID
			}
			sender := tools.SenderClient(ctx)
			sendMessage, err := sender.SendMessage(&message, &clientMessage)
			if err != nil {
				log.Errorf("send message error %v", zap.Error(err))
			}
			log.Infof("新消息ID %s", sendMessage)
		}
	})
	if err := testBot.Run(ctx); err != nil {
		panic(err)
	}
}
