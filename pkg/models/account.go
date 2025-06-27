package models

// Guild 群聊
type Guild struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GuildMember struct {
	*User `json:",inline"`

	GuildName string   `json:"alias"`
	GuildRole []string `json:"role"`
}

type User struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
