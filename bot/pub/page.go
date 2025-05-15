package pub

// PagedResponse 通用分页响应结构
type PagedResponse[T any] struct {
	Data []T    `json:"data"`           // 数据列表
	Next string `json:"next,omitempty"` // 分页令牌（可选）
}
