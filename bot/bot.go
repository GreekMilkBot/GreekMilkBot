package bot

type Bot struct {
	SelfID string
}

func NewBot(selfID string) *Bot {
	return &Bot{
		SelfID: selfID,
	}
}
