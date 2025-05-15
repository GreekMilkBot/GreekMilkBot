package pub

// Channel 类型定义
type Channel struct {
	ID       string      `json:"id"`
	Type     ChannelType `json:"type"`
	Name     string      `json:"name,omitempty"`
	ParentID string      `json:"parent_id,omitempty"`
}

// ChannelType 枚举定义
type ChannelType int

const (
	ChannelTypeText     ChannelType = 0 // 文本频道
	ChannelTypeDirect   ChannelType = 1 // 私聊频道
	ChannelTypeCategory ChannelType = 2 // 分类频道
	ChannelTypeVoice    ChannelType = 3 // 语音频道
)

func (ct ChannelType) String() string {
	switch ct {
	case ChannelTypeText:
		return "文本频道"
	case ChannelTypeDirect:
		return "私聊频道"
	case ChannelTypeCategory:
		return "分类频道"
	case ChannelTypeVoice:
		return "语音频道"
	default:
		return ""
	}
}

// API 请求结构体
type (
	GetChannelRequest struct {
		ChannelID string `json:"channel_id"`
	}

	ListChannelsRequest struct {
		GuildID string `json:"guild_id"`
		Next    string `json:"next,omitempty"`
	}

	CreateChannelRequest struct {
		GuildID string  `json:"guild_id"`
		Data    Channel `json:"data"`
	}

	UpdateChannelRequest struct {
		ChannelID string  `json:"channel_id"`
		Data      Channel `json:"data"`
	}

	DeleteChannelRequest struct {
		ChannelID string `json:"channel_id"`
	}

	MuteChannelRequest struct {
		ChannelID string `json:"channel_id"`
		Duration  int64  `json:"duration"`
	}

	CreateUserChannelRequest struct {
		UserID  string `json:"user_id"`
		GuildID string `json:"guild_id,omitempty"`
	}
)

// 分页响应结构体
type ChannelListResponse struct {
	Items     []Channel `json:"list"`
	NextToken string    `json:"next,omitempty"`
}

// ChannelService 频道服务接口定义
type ChannelService interface {
	GetChannel(req *GetChannelRequest) (*Channel, error)
	ListChannels(req *ListChannelsRequest) (*ChannelListResponse, error)
	CreateChannel(req *CreateChannelRequest) (*Channel, error)
	UpdateChannel(req *UpdateChannelRequest) (*Channel, error)
	DeleteChannel(req *DeleteChannelRequest) error
	MuteChannel(req *MuteChannelRequest) error
	CreateUserChannel(req *CreateUserChannelRequest) (*Channel, error)
}
