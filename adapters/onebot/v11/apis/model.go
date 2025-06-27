package apis

import (
	"encoding/json"
	"reflect"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"
)

type OneBotCustomContent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func (o OneBotCustomContent) String() string {
	marshal, _ := json.Marshal(&o)
	return string(marshal)
}

func init() {
	models.RegisterContent("onebot11_custom", reflect.TypeOf((*OneBotCustomContent)(nil)))
}
