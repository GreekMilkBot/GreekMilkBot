package GreekMilkBot

import (
	"context"
	"slices"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"
)

type BotContext struct {
	context.Context
	BotID int
	Tools []string

	GMBSender
	models.ResourceProviderFinderImpl
}

func (b BotContext) ToolAvailable(ID string) bool {
	return slices.Contains(b.Tools, ID)
}
