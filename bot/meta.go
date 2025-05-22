package bot

import (
	"encoding/json"
	"errors"
	"time"
)

type Channel struct {
	Type string `json:"type"`
}

func (c *Channel) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m["channel"] == nil {
		return errors.New("channel is required")
	}
	if child, ok := m["channel"].(map[string]interface{}); ok {
		switch child["type"] {
		case "private":
			chatData, err := json.Marshal(child)
			if err != nil {
				return err
			}
			chat := &Chat{}
			if err := json.Unmarshal(chatData, chat); err != nil {
				return err
			}
			*c = chat.Channel
			return nil
		case "guild":
			guildData, err := json.Marshal(child)
			if err != nil {
				return err
			}
			guild := &Guild{}
			if err := json.Unmarshal(guildData, guild); err != nil {
				return err
			}
			*c = guild.Channel
			return nil
		}
	}
	return errors.New("unknown channel type")
}

// Chat  私聊
type Chat struct {
	Channel `json:"channel"`
	Target  User `json:"user"`
}

// Guild 群聊
type Guild struct {
	Channel `json:"channel"`
	Id      string `json:"id"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
}

type GuildMember struct {
	User `json:"user"`

	GuildName   string `json:"name"`
	GuildAvatar string `json:"avatar"`

	GuildJoinedAt time.Time `json:"joined_at"`
}

type User struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
