package bot

type PacketType string

var (
	PacketMessage = PacketType("msg") // 消息
	PacketAction  = PacketType("act") // 控制

)

type Packet struct {
	Plugin string
	Type   PacketType

	Data any
}
