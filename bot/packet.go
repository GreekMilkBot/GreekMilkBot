package bot

type PacketType string

var (
	PacketMessage = PacketType("msg")  // 消息
	PacketAction  = PacketType("act")  // 控制
	PacketMeta    = PacketType("meta") // 元数据

)

type Packet struct {
	Plugin string
	Type   PacketType

	Data any
}
