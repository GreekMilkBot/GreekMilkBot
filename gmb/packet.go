package gmb

type PacketType string

var (
	PacketMessage = PacketType("msg") // 消息
	PacketAction  = PacketType("act") // 控制
)

type Packet struct {
	plugin string
	pType  PacketType

	data any
}
