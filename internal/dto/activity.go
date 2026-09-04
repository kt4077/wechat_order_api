package dto

import "activity/internal/model"

type ActivitySaveReq struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Cover         string            `json:"cover"`
	Category      string            `json:"category"`
	Type          int8              `json:"type"`
	Province      string            `json:"province"`
	City          string            `json:"city"`
	District      string            `json:"district"`
	Address       string            `json:"address"`
	Latitude      float64           `json:"latitude"`
	Longitude     float64           `json:"longitude"`
	PoiName       string            `json:"poi_name"`
	Description   string            `json:"description"`
	Notice        string            `json:"notice"`
	SignStartTime string            `json:"sign_start_time"`
	SignEndTime   string            `json:"sign_end_time"`
	StartTime     string            `json:"start_time"`
	EndTime       string            `json:"end_time"`
	Quota         int               `json:"quota"`
	Organizer     string            `json:"organizer"`
	ContactName   string            `json:"contact_name"`
	ContactPhone  string            `json:"contact_phone"`
	AutoAudit     int8              `json:"auto_audit"`
	Status        int8              `json:"status"`
	FormConfig    []model.FormField `json:"form_config"`
}

type ActivityQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Category string `json:"category" form:"category"`
	Type     int    `json:"type" form:"type"`
	Status   string `json:"status" form:"status"`
	Official int    `json:"official" form:"official"`
	Scope    string `json:"scope" form:"scope"`
	UserID   int64  `json:"-" form:"-"`
}

type ActivityListItem struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Cover         string  `json:"cover"`
	Category      string  `json:"category"`
	Type          int8    `json:"type"`
	Province      string  `json:"province"`
	City          string  `json:"city"`
	District      string  `json:"district"`
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	PoiName       string  `json:"poi_name"`
	SignStartTime string  `json:"sign_start_time"`
	SignEndTime   string  `json:"sign_end_time"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Quota         int     `json:"quota"`
	SignedCount   int     `json:"signed_count"`
	PassCount     int     `json:"pass_count"`
	RemainQuota   int     `json:"remain_quota"`
	Status        int8    `json:"status"`
	StatusText    string  `json:"status_text"`
	IsOfficial    int8    `json:"is_official"`
	CreatorID     int64   `json:"creator_id"`
	CreatorName   string  `json:"creator_name"`
	Organizer     string  `json:"organizer"`
	ViewCount     int     `json:"view_count"`
	PendingCount  int     `json:"pending_count"`
	CreateTime    string  `json:"create_time"`
}

type ActivityDetail struct {
	ActivityListItem
	Description string            `json:"description"`
	Notice      string            `json:"notice"`
	FormConfig  []model.FormField `json:"form_config"`
	ContactName string            `json:"contact_name"`
	AutoAudit   int8              `json:"auto_audit"`
	CanSignup   bool              `json:"can_signup"`
	SignupTip   string            `json:"signup_tip"`
	MySignupID  int64             `json:"my_signup_id"`
	MyStatus    int8              `json:"my_status"`
	CanManage   bool              `json:"can_manage"`
}

type ActivityStatusReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}

type ActivityAuditReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}

type ActivityCopyReq struct {
	ID int64 `json:"id"`
}
