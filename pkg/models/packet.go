package models

import "time"

type PacketType string

var (
	PacketMessage = PacketType("msg")  // 消息
	PacketAction  = PacketType("act")  // 控制
	PacketMeta    = PacketType("meta") // 元数据

)

type Packet struct {
	Plugin int
	Type   PacketType

	Data any
}

type Meta struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Message struct {
	ID    string       `json:"id"`
	Owner *GuildMember `json:"user"`

	MsgType string `json:"type"`
	Guild   *Guild `json:"guild"`

	Quote   *Message  `json:"quote,omitempty"`
	Content Contents  `json:"content"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Event struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data,omitempty"`
}
