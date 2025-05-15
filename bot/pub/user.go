package pub

// User 定义用户信息结构体
type User struct {
	ID     string  `json:"id"`               // 用户ID
	Name   *string `json:"name,omitempty"`   // 用户名（可选，展示优先级低于Nick）
	Nick   *string `json:"nick,omitempty"`   // 用户昵称（优先展示）
	Avatar *string `json:"avatar,omitempty"` // 用户头像链接（可选）
	IsBot  *bool   `json:"is_bot,omitempty"` // 是否机器人标识（可选）
}

// FriendApproveRequest 处理好友申请请求参数
type FriendApproveRequest struct {
	MessageID string  `json:"message_id"`        // 请求ID
	Approve   bool    `json:"approve"`           // 是否通过请求
	Comment   *string `json:"comment,omitempty"` // 备注信息（可选）
}

// UserClient 用户服务接口定义
type UserClient interface {
	// GetUser 获取单个用户信息
	GetUser(userID string) (*User, error)

	// ListFriends 获取好友列表
	ListFriends(nextToken string) (*PagedResponse[User], error)

	// ApproveFriend 处理好友申请
	ApproveFriend(req *FriendApproveRequest) error
}
