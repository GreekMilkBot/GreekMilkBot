package bot

import (
	"fmt"
	"reflect"
)

type ContentText struct {
	Text string `json:"text"`
}

func (c ContentText) String() string {
	return c.Text
}

type ContentAt struct {
	Uid  string `json:"uid"`
	User *User  `json:"user"`
}

func (c ContentAt) String() string {
	return fmt.Sprintf("@%s", c.Uid)
}

type ContentImage struct {
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

func NewBase64ContentImage(mediaType, data, summary string) ContentImage {
	return ContentImage{
		URL:     fmt.Sprintf("base64://%s?ContentType=%s", data, mediaType),
		Summary: summary,
	}
}

func (c ContentImage) String() string {
	return fmt.Sprintf("image[summary=%s,blob]", c.Summary)
}

func init() {
	RegisterContent("text", reflect.TypeOf((*ContentText)(nil)))
	RegisterContent("at", reflect.TypeOf((*ContentAt)(nil)))
	RegisterContent("image", reflect.TypeOf((*ContentImage)(nil)))
}
