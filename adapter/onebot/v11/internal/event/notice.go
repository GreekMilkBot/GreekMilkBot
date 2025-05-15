package event

type NoticeEvent struct {
	BaseEvent
	NoticeType string `json:"notice_type"`
	TargetID   int64  `json:"target_id"`
}
