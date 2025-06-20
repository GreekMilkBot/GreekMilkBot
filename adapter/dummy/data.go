package dummy

import "github.com/GreekMilkBot/GreekMilkBot/bot"

type Tree struct {
	Owner string `json:"owner"`

	Users   map[string]*bot.User `json:"user"`
	Private []*Private           `json:"private"`
	Guilds  map[string]*Guild    `json:"guild"`

	Messages map[string]*Message `json:"messages"`
}
type Guild struct {
	bot.Guild `json:",inline"`
	Users     []string `json:"users"`
}

type Private struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

type Message struct {
	Sender  bot.User      `json:"sender"`
	Content *bot.Contents `json:"messages"`
}
