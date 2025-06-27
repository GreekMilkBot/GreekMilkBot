package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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

type ContentUnknown struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (c ContentUnknown) String() string {
	return fmt.Sprintf("unknown[type=%s]", c.Type)
}
