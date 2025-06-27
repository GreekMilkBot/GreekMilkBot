package event

import "encoding/json"

type ActionEvent struct {
	BaseEvent
	Status  string `json:"status"`
	Code    int    `json:"retcode"`
	Data    string `json:"-"`
	Message string `json:"message"`
}

func (a *ActionEvent) UnmarshalJSON(bytes []byte) error {
	type Alias ActionEvent
	var out Alias
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	marshal, err := json.Marshal(m["data"])
	if err != nil {
		return err
	}
	out.Data = string(marshal)
	*a = ActionEvent(out)
	return nil
}
