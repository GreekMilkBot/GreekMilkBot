package bot

import (
	"time"
)

// Guild 群聊
type Guild struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GuildMember struct {
	User `json:"user"`

	GuildName string `json:"name"`

	GuildJoinedAt time.Time `json:"joined_at"`
}

type User struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
