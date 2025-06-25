package bot

type ClientMessage struct {
	QuoteID string    `json:"quote_id"`
	Message *Contents `json:"contents"`
}
