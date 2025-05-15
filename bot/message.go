package bot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	covertMap  = make(map[string]reflect.Type)
	covertMapR = make(map[reflect.Type]string)
)

type (
	Contents    []Content
	RAWContents []RawContent
)

func (contents Contents) String() string {
	var data []string
	for _, content := range contents {
		data = append(data, content.String())
	}
	return strings.Join(data, "")
}

func (contents Contents) ToRAWContents() (RAWContents, error) {
	result := make(RAWContents, 0, len(contents))
	for _, content := range contents {
		t := reflect.TypeOf(content)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		s := covertMapR[t]
		if s == "" {
			return nil, fmt.Errorf("unknown type %T", content)
		}
		data, err := json.Marshal(content)
		if err != nil {
			return nil, err
		}
		result = append(result, RawContent{
			Type: s,
			Data: string(data),
		})
	}
	return result, nil
}

func (contents RAWContents) ToContents() (Contents, error) {
	result := make(Contents, 0, len(contents))
	for _, content := range contents {
		r := covertMap[content.Type]
		if r == nil {
			return nil, fmt.Errorf("unknown content type: %s", content.Type)
		}
		i := reflect.New(r).Interface()
		if err := json.Unmarshal([]byte(content.Data), i); err != nil {
			return nil, err
		}
		result = append(result, reflect.ValueOf(i).Elem().Interface().(Content))
	}
	return result, nil
}

type Content interface {
	fmt.Stringer
}
type Message struct {
	ID    string `json:"id"`
	Owner *User  `json:"user"`

	MsgType string `json:"type"`
	Guild   *Guild `json:"guild"`

	Quote      *Message    `json:"quote,omitempty"`
	Content    Contents    `json:"-"`
	RawContent RAWContents `json:"content"`
	Created    time.Time   `json:"created"`
	Updated    time.Time   `json:"updated"`
}

func (m *Message) UnmarshalJSON(bytes []byte) error {
	type Alias Message
	var alias Alias
	var err error
	if err := json.Unmarshal(bytes, &alias); err != nil {
		return err
	}
	alias.Content, err = alias.RawContent.ToContents()
	if err != nil {
		return err
	}
	*m = Message(alias)
	return nil
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	alias := (*Alias)(m)
	var err error
	alias.RawContent, err = m.Content.ToRAWContents()
	if err != nil {
		return nil, err
	}
	// 临时清空 Content 字段以避免循环引用
	originalContent := alias.Content
	alias.Content = nil
	defer func() { alias.Content = originalContent }()
	return json.Marshal(alias)
}

type RawContent struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var baseType = reflect.TypeOf((*Content)(nil)).Elem()

func RegisterContent(key string, typeOf reflect.Type) {
	if _, ok := covertMap[key]; ok {
		panic("duplicate key " + key)
	}
	if !typeOf.Implements(baseType) {
		panic("type" + key + " not implements Content")
	}
	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	covertMap[key] = typeOf
	covertMapR[typeOf] = key
}

type ContentText struct {
	Text string `json:"text"`
}

func (c ContentText) String() string {
	return c.Text
}

type ContentAt struct {
	Uid string `json:"uid"`
}

func (c ContentAt) String() string {
	return fmt.Sprintf("@%s", c.Uid)
}

func init() {
	RegisterContent("text", reflect.TypeOf((*ContentText)(nil)))
	RegisterContent("at", reflect.TypeOf((*ContentAt)(nil)))
}
