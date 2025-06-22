package dummy

import (
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"net/http"
	"time"
)

type Tree struct {
	*http.ServeMux `json:"-"`

	Owner string `json:"owner"`

	Users    map[string]*bot.User `json:"user"`
	Guilds   map[string]*Guild    `json:"guild"`
	Sessions map[string]*Session  `json:"session"`
	Messages map[string]*Message  `json:"messages"`
}

func NewTree() *Tree {
	t := new(Tree)
	return t
}

type Session struct {
	SType  string `json:"type"` // private or group
	Target string `json:"target"`
}

type Guild struct {
	bot.Guild `json:",inline"`
	Users     []*GroupUser `json:"users"` // user id
}

type GroupUser struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type Message struct {
	Sender   bot.User         `json:"sender"`
	CreateAt time.Time        `json:"create_at"`
	Content  *bot.RAWContents `json:"messages"`
}
