package model

import "time"

// 用户角色常量（从 1 开始，避免使用 0 作为状态值）
const (
	RoleNormal     int8 = 1 // 普通用户
	RoleMerchant   int8 = 2 // 入驻管理员（主办方）
	RoleSuperAdmin int8 = 3 // 超级管理员
)

// 用户状态常量
const (
	UserStatusNormal  int8 = 1 // 正常
	UserStatusDisable int8 = 2 // 禁用
)

// 性别常量（从 1 开始，避免使用 0 作为枚举值）
const (
	GenderUnknown int8 = 1 // 未知
	GenderMale    int8 = 2 // 男
	GenderFemale  int8 = 3 // 女
)

// User 微信端用户表
type User struct {
	BaseModel
	Openid        string     `gorm:"column:openid;type:varchar(64);uniqueIndex:uk_openid;not null;comment:微信openid" json:"openid"`
	Unionid       string     `gorm:"column:unionid;type:varchar(64);index;comment:微信unionid" json:"unionid"`
	Nickname      string     `gorm:"column:nickname;type:varchar(64);comment:用户昵称" json:"nickname"`
	Avatar        string     `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`
	Phone         string     `gorm:"column:phone;type:varchar(255);comment:手机号（AES加密存储）" json:"-"`
	RealName      string     `gorm:"column:real_name;type:varchar(64);comment:真实姓名" json:"real_name"`
	Gender        int8       `gorm:"column:gender;default:1;comment:性别 1未知 2男 3女" json:"gender"`
	Role          int8       `gorm:"column:role;default:1;index;comment:角色 1普通用户 2入驻管理员 3超级管理员" json:"role"`
	Status        int8       `gorm:"column:status;default:1;comment:状态 1正常 2禁用" json:"status"`
	MerchantFlag  int8       `gorm:"column:merchant_flag;default:1;comment:入驻权限 1无 2已开通" json:"merchant_flag"`
	LastLoginTime *time.Time `gorm:"column:last_login_time;comment:最近登录时间" json:"last_login_time"`
}

// TableName 数据表名
func (User) TableName() string { return "user" }
