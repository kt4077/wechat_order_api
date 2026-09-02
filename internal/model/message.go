package model

// 消息类型（从 1 开始，避免使用 0 作为状态值）
const (
	MsgTypeSystem   int8 = 1 // 系统通知
	MsgTypeSignup   int8 = 2 // 报名相关
	MsgTypeAudit    int8 = 3 // 报名审核
	MsgTypeMerchant int8 = 4 // 入驻审核
	MsgTypeActivity int8 = 5 // 活动变更
)

// Message 站内消息表（与微信订阅消息互补，保证小程序内可查历史通知）
type Message struct {
	BaseModel
	UserID    int64  `gorm:"column:user_id;type:bigint;not null;index:idx_message_user;comment:接收用户ID" json:"user_id"`
	Type      int8   `gorm:"column:type;default:1;comment:消息类型 1系统 2报名 3报名审核 4入驻审核 5活动变更" json:"type"`
	Title     string `gorm:"column:title;type:varchar(120);comment:消息标题" json:"title"`
	Content   string `gorm:"column:content;type:varchar(1000);comment:消息内容" json:"content"`
	RelatedID int64  `gorm:"column:related_id;type:bigint;default:0;comment:关联业务ID" json:"related_id"`
	IsRead    int8   `gorm:"column:is_read;default:1;comment:是否已读 1否 2是" json:"is_read"`
}

// TableName 数据表名
func (Message) TableName() string { return "message" }
