package event

type NoticeEvent map[string]any

func (n NoticeEvent) GetSelfId() uint64 {
	return uint64(n["self_id"].(int))
}

func (n NoticeEvent) GetNoticeType() string {
	return n["notice_type"].(string)
}

func (n NoticeEvent) GetType() EventType {
	return EventTypeNotice
}
