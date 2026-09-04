package model

import "time"

const (
	AdminRoleSuper int8 = 1
	AdminRoleSub   int8 = 2
)

type Admin struct {
	BaseModel
	Username      string     `gorm:"column:username;type:varchar(64);uniqueIndex:uk_admin_username;not null;comment:登录账号" json:"username"`
	Password      string     `gorm:"column:password;type:varchar(128);not null;comment:bcrypt密码" json:"-"`
	Nickname      string     `gorm:"column:nickname;type:varchar(64);comment:昵称" json:"nickname"`
	Avatar        string     `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`
	Role          int8       `gorm:"column:role;default:2;comment:角色 1超级管理员 2子管理员" json:"role"`
	Status        int8       `gorm:"column:status;default:1;comment:状态 1正常 2禁用" json:"status"`
	Permissions   string     `gorm:"column:permissions;type:text;comment:权限标识JSON数组（超管忽略）" json:"permissions"`
	LastLoginTime *time.Time `gorm:"column:last_login_time;comment:最近登录时间" json:"last_login_time"`
	LastLoginIP   string     `gorm:"column:last_login_ip;type:varchar(64);comment:最近登录IP" json:"last_login_ip"`
}

func (Admin) TableName() string { return "admin" }

type SysConfig struct {
	BaseModel
	ConfigKey   string `gorm:"column:config_key;type:varchar(64);uniqueIndex:uk_config_key;not null;comment:配置键" json:"config_key"`
	ConfigValue string `gorm:"column:config_value;type:text;comment:配置值" json:"config_value"`
	Remark      string `gorm:"column:remark;type:varchar(255);comment:配置说明" json:"remark"`
}

func (SysConfig) TableName() string { return "sys_config" }

type OperationLog struct {
	BaseModel
	AdminID   int64  `gorm:"column:admin_id;type:bigint;index:idx_log_admin;comment:操作人ID" json:"admin_id"`
	AdminName string `gorm:"column:admin_name;type:varchar(64);comment:操作人名称" json:"admin_name"`
	Module    string `gorm:"column:module;type:varchar(64);comment:操作模块" json:"module"`
	Action    string `gorm:"column:action;type:varchar(64);comment:操作动作" json:"action"`
	Detail    string `gorm:"column:detail;type:varchar(1000);comment:操作详情" json:"detail"`
	IP        string `gorm:"column:ip;type:varchar(64);comment:操作IP" json:"ip"`
}

func (OperationLog) TableName() string { return "operation_log" }
