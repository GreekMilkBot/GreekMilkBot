package models

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
	Resource Resource `json:"data"`
	Summary  string   `json:"summary"`
}

func (c ContentImage) String() string {
	return fmt.Sprintf("image[summary=%s,blob]", c.Summary)
}

func init() {
	RegisterContent("text", reflect.TypeOf((*ContentText)(nil)))
	RegisterContent("at", reflect.TypeOf((*ContentAt)(nil)))
	RegisterContent("image", reflect.TypeOf((*ContentImage)(nil)))
}
