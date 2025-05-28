package bot

type EventType string

var (
	EventTypeMessage = EventType("msg") // 消息
	EventTypeAction  = EventType("act") // 控制
)

type Event interface {
	Type() EventType
}
