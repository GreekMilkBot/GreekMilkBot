package message

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) OnMessage(msg string) {

}
