package model

import "time"

// 入驻申请状态（从 1 开始，避免使用 0 作为状态值）
const (
	ApplyStatusPending  int8 = 1 // 待审核
	ApplyStatusApproved int8 = 2 // 审核通过
	ApplyStatusRejected int8 = 3 // 审核驳回
)

// 入驻类型
const (
	MerchantTypePerson int8 = 1 // 个人
	MerchantTypeOrg    int8 = 2 // 机构
)

// MerchantApply 用户入驻（成为主办方）申请表
type MerchantApply struct {
	BaseModel
	UserID              int64      `gorm:"column:user_id;type:bigint;not null;index:idx_apply_user;comment:申请人ID" json:"user_id"`
	Nickname            string     `gorm:"column:nickname;type:varchar(64);comment:申请人昵称快照" json:"nickname"`
	Avatar              string     `gorm:"column:avatar;type:varchar(255);comment:申请人头像快照" json:"avatar"`
	Type                int8       `gorm:"column:type;default:1;comment:入驻类型 1个人 2机构" json:"type"`
	ContactName         string     `gorm:"column:contact_name;type:varchar(64);not null;comment:联系人姓名" json:"contact_name"`
	ContactPhone        string     `gorm:"column:contact_phone;type:varchar(255);not null;comment:联系电话（AES加密）" json:"-"`
	Intro               string     `gorm:"column:intro;type:varchar(1000);comment:入驻简介/申请说明" json:"intro"`
	QualificationImages string     `gorm:"column:qualification_images;type:text;comment:资质图片JSON数组" json:"qualification_images"`
	Status              int8       `gorm:"column:status;default:1;index:idx_apply_status;comment:状态 1待审核 2通过 3驳回" json:"status"`
	AuditRemark         string     `gorm:"column:audit_remark;type:varchar(500);comment:审核备注" json:"audit_remark"`
	AuditorID           int64      `gorm:"column:auditor_id;type:bigint;default:0;comment:审核人ID" json:"auditor_id"`
	AuditorName         string     `gorm:"column:auditor_name;type:varchar(64);comment:审核人名称" json:"auditor_name"`
	AuditTime           *time.Time `gorm:"column:audit_time;comment:审核时间" json:"audit_time"`
	ClientIP            string     `gorm:"column:client_ip;type:varchar(64);comment:提交IP" json:"client_ip"`
}

// TableName 数据表名
func (MerchantApply) TableName() string { return "merchant_apply" }
