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

func (contents *Contents) UnmarshalJSON(bytes []byte) error {
	var raw RAWContents
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return err
	}
	result := make(Contents, 0, len(raw))
	for _, content := range raw {
		r := covertMap[content.Type]
		if r == nil {
			result = append(result,
				ContentUnknown{
					Type:  content.Type,
					Value: content.Data,
				})
			continue
		}
		i := reflect.New(r).Interface()
		if err := json.Unmarshal([]byte(content.Data), i); err != nil {
			return err
		}
		result = append(result, reflect.ValueOf(i).Elem().Interface().(Content))
	}
	*contents = result
	return nil
}

func (contents *Contents) MarshalJSON() ([]byte, error) {
	result := make(RAWContents, 0, len(*contents))
	for _, content := range *contents {
		t := reflect.TypeOf(content)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		s := covertMapR[t]
		if s == "" {
			switch typedContent := content.(type) {
			case ContentUnknown:
				result = append(result, RawContent{
					Type: typedContent.Type,
					Data: typedContent.Value,
				})
				continue
			default:
				return nil, fmt.Errorf("unknown type %T", typedContent)
			}
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
	return json.Marshal(result)
}

func (contents *Contents) String() string {
	var data []string
	for _, content := range *contents {
		data = append(data, content.String())
	}
	return strings.Join(data, "")
}

type Content interface {
	fmt.Stringer
}
type Message struct {
	ID    string `json:"id"`
	Owner *User  `json:"user"`

	MsgType string `json:"type"`
	Guild   *Guild `json:"guild"`

	Quote   *Message  `json:"quote,omitempty"`
	Content Contents  `json:"content"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
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

type ContentUnknown struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (c ContentUnknown) String() string {
	return fmt.Sprintf("unknown[type=%s]", c.Type)
}

func init() {
	RegisterContent("text", reflect.TypeOf((*ContentText)(nil)))
	RegisterContent("at", reflect.TypeOf((*ContentAt)(nil)))
	RegisterContent("image", reflect.TypeOf((*ContentImage)(nil)))
}
