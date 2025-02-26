package bot

type Bot struct {
	SelfID int64
}

func NewBot(selfID int64) *Bot {
	return &Bot{
		SelfID: selfID,
	}
}
