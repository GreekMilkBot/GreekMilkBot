package models

type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

type CommonMessage struct {
	Time        int64       `json:"time"`
	MessageType MessageType `json:"message_type"`
	SubType     string      `json:"sub_type"`
	MessageID   int         `json:"message_id"`
	UserId      uint64      `json:"user_id"`
	Message     []Message   `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Font        int         `json:"font"`
	Sender      Sender      `json:"sender"`

	// group only
	GroupID   uint64    `json:"group_id"`
	Anonymous Anonymous `json:"anonymous"`
}

type Message struct {
	MsgType string         `json:"type"`
	MsgData map[string]any `json:"data"`
}

type SexType string

const (
	SexTypeMale    SexType = "male"
	SexTypeFemale  SexType = "female"
	SexTypeUnknown SexType = "unknown"
)

type Sender struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
	Area     string `json:"area"`
	Level    string `json:"level"`
	Role     string `json:"role"`
	Title    string `json:"title"`
}

type Anonymous struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}
