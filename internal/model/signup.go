package model

import "time"

const (
	SignupStatusPending  int8 = 1
	SignupStatusApproved int8 = 2
	SignupStatusRejected int8 = 3
	SignupStatusCanceled int8 = 4
)

type Signup struct {
	BaseModel
	ActivityID    int64      `gorm:"column:activity_id;type:bigint;not null;index:idx_signup_activity;comment:活动ID" json:"activity_id"`
	ActivityTitle string     `gorm:"column:activity_title;type:varchar(120);comment:活动名称快照" json:"activity_title"`
	ActivityCover string     `gorm:"column:activity_cover;type:varchar(255);comment:活动封面快照" json:"activity_cover"`
	ActivityType  int8       `gorm:"column:activity_type;default:1;comment:活动形式快照" json:"activity_type"`
	StartTime     *time.Time `gorm:"column:start_time;comment:活动开始时间快照" json:"start_time"`
	EndTime       *time.Time `gorm:"column:end_time;comment:活动结束时间快照" json:"end_time"`
	Address       string     `gorm:"column:address;type:varchar(255);comment:活动地址快照" json:"address"`
	UserID        int64      `gorm:"column:user_id;type:bigint;not null;index:idx_signup_user;comment:报名用户ID" json:"user_id"`
	Nickname      string     `gorm:"column:nickname;type:varchar(64);comment:报名人昵称" json:"nickname"`
	Avatar        string     `gorm:"column:avatar;type:varchar(255);comment:报名人头像" json:"avatar"`
	FormData      string     `gorm:"column:form_data;type:text;comment:报名表单数据JSON" json:"form_data"`
	Status        int8       `gorm:"column:status;default:1;index:idx_signup_status;comment:状态 1待审核 2通过 3驳回 4已撤销" json:"status"`
	AuditRemark   string     `gorm:"column:audit_remark;type:varchar(500);comment:审核备注" json:"audit_remark"`
	AuditorID     int64      `gorm:"column:auditor_id;type:bigint;default:0;comment:审核人ID" json:"auditor_id"`
	AuditorName   string     `gorm:"column:auditor_name;type:varchar(64);comment:审核人名称" json:"auditor_name"`
	AuditTime     *time.Time `gorm:"column:audit_time;comment:审核时间" json:"audit_time"`
	AppointTime   *time.Time `gorm:"column:appoint_time;comment:报名者填写的预约到场时间" json:"appoint_time"`
	CancelTime    *time.Time `gorm:"column:cancel_time;comment:撤销时间" json:"cancel_time"`
	ClientIP      string     `gorm:"column:client_ip;type:varchar(64);comment:提交IP" json:"client_ip"`
}

func (Signup) TableName() string { return "signup" }
