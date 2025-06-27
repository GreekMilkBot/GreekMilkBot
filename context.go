package GreekMilkBot

import (
	"context"
	"slices"
)

type BotContext struct {
	context.Context
	BotID string
	Tools []string

	GMBSender
}

func (b BotContext) ToolAvailable(ID string) bool {
	return slices.Contains(b.Tools, ID)
}
