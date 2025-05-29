package bot

import (
	"encoding/json"
	"time"
)

type MsgContent interface {
}
type Message struct {
	ID    string `json:"id"`
	Owner *User  `json:"user"`

	MsgType string `json:"type"`
	Guild   *Guild `json:"guild"`

	Quote      *Message     `json:"quote,omitempty"`
	Content    []MsgContent `json:"-"`
	ContentRaw []Content    `json:"content"`
	Created    time.Time    `json:"created"`
	Updated    time.Time    `json:"updated"`
}

func (m *Message) UnmarshalJSON(bytes []byte) error {
	type Alias Message
	var alias Alias
	if err := json.Unmarshal(bytes, &alias); err != nil {
		return err
	}
	for _, content := range alias.ContentRaw {
		switch content.Type {
		case "text":
			var ct ContentText
			if err := json.Unmarshal([]byte(content.Data), &ct); err != nil {
				return err
			}
			alias.Content = append(alias.Content, ct)
		}
	}
	*m = Message(alias)
	return nil
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	alias := (*Alias)(m)

	// 将 Content 转换回 ContentRaw
	for _, content := range m.Content {
		switch c := content.(type) {
		case ContentText:
			data, err := json.Marshal(c)
			if err != nil {
				return nil, err
			}
			alias.ContentRaw = append(alias.ContentRaw, Content{
				Type: "text",
				Data: string(data),
			})
		}
	}

	// 临时清空 Content 字段以避免循环引用
	originalContent := alias.Content
	alias.Content = nil
	defer func() { alias.Content = originalContent }()

	return json.Marshal(alias)
}

type Content struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type ContentText struct {
	Text string `json:"text"`
}
