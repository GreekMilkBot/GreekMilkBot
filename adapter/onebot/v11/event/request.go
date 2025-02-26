package event

type RequestEvent struct {
	BaseEvent
	RequestType string `json:"request_type"`
	UserID      int64  `json:"user_id"`
}
