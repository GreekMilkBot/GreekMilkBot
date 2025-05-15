package bot

type Event struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}
