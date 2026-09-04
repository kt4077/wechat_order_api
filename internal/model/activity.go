package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ActivityStatusDraft    int8 = 1
	ActivityStatusNotStart int8 = 2
	ActivityStatusSigning  int8 = 3
	ActivityStatusEnded    int8 = 4
	ActivityStatusOffline  int8 = 5
	ActivityStatusPending  int8 = 6
	ActivityStatusRejected int8 = 7
)

const (
	ActivityTypeOnline  int8 = 1
	ActivityTypeOffline int8 = 2
)

const (
	ActivityOfficialNo  int8 = 1
	ActivityOfficialYes int8 = 2
)

type Activity struct {
	BaseModel
	Title         string     `gorm:"column:title;type:varchar(120);not null;comment:活动名称" json:"title"`
	Cover         string     `gorm:"column:cover;type:varchar(255);comment:活动封面/海报" json:"cover"`
	Category      string     `gorm:"column:category;type:varchar(32);index;comment:活动分类" json:"category"`
	Type          int8       `gorm:"column:type;default:1;comment:活动形式 1线上 2线下" json:"type"`
	Province      string     `gorm:"column:province;type:varchar(32);comment:省份" json:"province"`
	City          string     `gorm:"column:city;type:varchar(32);comment:城市" json:"city"`
	District      string     `gorm:"column:district;type:varchar(32);comment:区/县" json:"district"`
	Address       string     `gorm:"column:address;type:varchar(255);comment:详细地址" json:"address"`
	Latitude      float64    `gorm:"column:latitude;type:decimal(10,7);default:0;comment:纬度" json:"latitude"`
	Longitude     float64    `gorm:"column:longitude;type:decimal(10,7);default:0;comment:经度" json:"longitude"`
	PoiName       string     `gorm:"column:poi_name;type:varchar(120);comment:地图点位名称" json:"poi_name"`
	Description   string     `gorm:"column:description;type:text;comment:活动介绍" json:"description"`
	Notice        string     `gorm:"column:notice;type:text;comment:报名须知" json:"notice"`
	SignStartTime *time.Time `gorm:"column:sign_start_time;comment:报名开始时间" json:"sign_start_time"`
	SignEndTime   *time.Time `gorm:"column:sign_end_time;comment:报名结束时间" json:"sign_end_time"`
	StartTime     *time.Time `gorm:"column:start_time;comment:活动开始时间" json:"start_time"`
	EndTime       *time.Time `gorm:"column:end_time;comment:活动结束时间" json:"end_time"`
	Quota         int        `gorm:"column:quota;default:0;comment:报名名额 0表示不限" json:"quota"`
	SignedCount   int        `gorm:"column:signed_count;default:0;comment:已报名数量" json:"signed_count"`
	PassCount     int        `gorm:"column:pass_count;default:0;comment:审核通过数量" json:"pass_count"`
	ViewCount     int        `gorm:"column:view_count;default:0;comment:浏览数" json:"view_count"`
	Organizer     string     `gorm:"column:organizer;type:varchar(120);comment:主办方名称" json:"organizer"`
	ContactName   string     `gorm:"column:contact_name;type:varchar(64);comment:主办方联系人" json:"contact_name"`
	ContactPhone  string     `gorm:"column:contact_phone;type:varchar(255);comment:主办方联系电话（AES加密）" json:"-"`
	AutoAudit     int8       `gorm:"column:auto_audit;default:1;comment:报名免审核 1否 2是" json:"auto_audit"`
	Status        int8       `gorm:"column:status;default:1;index:idx_activity_status;comment:状态 1草稿 2未开始 3报名中 4已结束 5已下架 6待审核 7审核驳回" json:"status"`
	IsOfficial    int8       `gorm:"column:is_official;default:1;index:idx_activity_official;comment:是否平台官方 1否 2是" json:"is_official"`
	CreatorID     int64      `gorm:"column:creator_id;type:bigint;not null;index:idx_activity_creator;comment:创建者ID" json:"creator_id"`
	CreatorName   string     `gorm:"column:creator_name;type:varchar(64);comment:创建者名称" json:"creator_name"`
	FormConfig    string     `gorm:"column:form_config;type:text;comment:自定义报名表单JSON配置" json:"form_config"`
	WarnNotified  int8       `gorm:"column:warn_notified;default:1;comment:名额预警已通知 1否 2是" json:"warn_notified"`
}

func (Activity) TableName() string { return "activity" }

type FormField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
	MinLen      int      `json:"min_len"`
	MaxLen      int      `json:"max_len"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	Sort        int      `json:"sort"`
	Default     string   `json:"default,omitempty"`
	Privacy     bool     `json:"privacy"`
}

func DefaultFormConfig() []FormField {
	return []FormField{
		{Key: "name", Label: "姓名", Type: "text", Required: true, Placeholder: "请输入姓名", MinLen: 2, MaxLen: 20, Sort: 1},
		{Key: "phone", Label: "手机号", Type: "phone", Required: true, Placeholder: "请输入11位手机号", Sort: 2, Privacy: true},
	}
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.FormConfig == "" {
		_ = tx
	}
	return nil
}
