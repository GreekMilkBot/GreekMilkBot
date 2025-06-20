package bot

type Event struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data,omitempty"`
}
