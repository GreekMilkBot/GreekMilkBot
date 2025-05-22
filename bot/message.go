package bot

import "time"

type Message struct {
	Owner   *User    `json:"user"`
	Channel *Channel `json:"guild"`

	Quote *Message `json:"quote,omitempty"`

	Content []Content `json:"content"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Content struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
