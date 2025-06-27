package core

type ActionResponse struct {
	ID       string `json:"id"`
	OK       bool   `json:"ok"`
	ErrorMsg string `json:"error,omitempty"`

	Data []string `json:"data,omitempty"`
}

type ActionRequest struct {
	ID     string   `json:"id"`
	Action string   `json:"action"`
	Params []string `json:"params,omitempty"`
}
