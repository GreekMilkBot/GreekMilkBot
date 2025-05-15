package pub

// Guild 群组信息
type Guild struct {
	ID     string  `json:"id"`     // 群组 ID
	Name   *string `json:"name"`   // 群组名称（可选）
	Avatar *string `json:"avatar"` // 群组头像（可选）
}

// GuildGetRequest 获取群组请求参数
type GuildGetRequest struct {
	GuildID string `json:"guild_id"` // 要获取的群组ID
}

// GuildListRequest 获取群组列表请求参数
type GuildListRequest struct {
	NextToken *string `json:"next,omitempty"` // 分页令牌（可选）
}

// GuildApproveRequest 处理群组邀请请求参数
type GuildApproveRequest struct {
	MessageID string  `json:"message_id"` // 请求ID
	Approve   bool    `json:"approve"`    // 是否通过请求
	Comment   *string `json:"comment"`    // 备注信息（可选）
}

// PageResult 分页结果
type PageResult struct {
	Items     []Guild `json:"items"` // 当前页数据
	NextToken *string `json:"next"`  // 下一页分页令牌（可选）
}

// GuildAPI 群组服务接口
type GuildAPI interface {
	GetGuild(req GuildGetRequest) (*Guild, error)
	ListGuilds(req GuildListRequest) (*PageResult, error)
	ApproveInvite(req GuildApproveRequest) error
}

// GuildMember 表示群组成员信息
type GuildMember struct {
	User     *User   `json:"user,omitempty"`   // 用户对象（可能不存在）
	Nick     *string `json:"nick,omitempty"`   // 用户在群组中的名称
	Avatar   *string `json:"avatar,omitempty"` // 用户在群组中的头像
	JoinedAt int64   `json:"joined_at"`        // 加入时间（Unix 时间戳）
}

// GuildMemberGetRequest 获取群组成员请求参数
type GuildMemberGetRequest struct {
	GuildID string `json:"guild_id"`
	UserID  string `json:"user_id"`
}

// GuildMemberListRequest 获取成员列表请求参数
type GuildMemberListRequest struct {
	GuildID string  `json:"guild_id"`
	Next    *string `json:"next,omitempty"` // 分页令牌
}

// GuildMemberListResponse 成员列表分页响应
type GuildMemberListResponse struct {
	Items     []GuildMember `json:"items"`
	NextToken string        `json:"next,omitempty"`
}

// GuildMemberKickRequest 踢出成员请求参数
type GuildMemberKickRequest struct {
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	Permanent *bool  `json:"permanent,omitempty"` // 是否永久踢出
}

// GuildMemberMuteRequest 禁言请求参数（实验性）
type GuildMemberMuteRequest struct {
	GuildID  string `json:"guild_id"`
	UserID   string `json:"user_id"`
	Duration int64  `json:"duration"` // 禁言时长（毫秒）
}

// GuildMemberApproveRequest 处理加群请求参数
type GuildMemberApproveRequest struct {
	MessageID string  `json:"message_id"`
	Approve   bool    `json:"approve"`
	Comment   *string `json:"comment,omitempty"` // 备注信息
}

type GuildMemberAPI interface {
	// Get 获取群成员信息
	Get(req GuildMemberGetRequest) (*GuildMember, error)
	// List 获取群成员列表
	List(req GuildMemberListRequest) (*GuildMemberListResponse, error)
	// Kick 踢出群组成员
	Kick(req GuildMemberKickRequest) error
	// Mute 禁言群组成员（实验性接口）
	Mute(req GuildMemberMuteRequest) error
	// Approve 处理加群请求
	Approve(req GuildMemberApproveRequest) error
}
