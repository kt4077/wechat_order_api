package dto

import "activity/internal/model"

type SignupSubmitReq struct {
	ActivityID  int64                  `json:"activity_id"`
	FormData    map[string]interface{} `json:"form_data"`
	AppointTime string                 `json:"appoint_time"`
}

type SignupAuditReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}

type SignupBatchAuditReq struct {
	IDs    []int64 `json:"ids"`
	Status int8    `json:"status"`
	Remark string  `json:"remark"`
}

type SignupQuery struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	ActivityID int64  `json:"activity_id" form:"activity_id"`
	Keyword    string `json:"keyword" form:"keyword"`
	Status     *int   `json:"status" form:"status"`
	Scope      string `json:"scope" form:"scope"`
	UserID     int64  `json:"-" form:"-"`
}

type SignupListItem struct {
	ID            int64                  `json:"id"`
	ActivityID    int64                  `json:"activity_id"`
	ActivityTitle string                 `json:"activity_title"`
	ActivityCover string                 `json:"activity_cover"`
	ActivityType  int8                   `json:"activity_type"`
	StartTime     string                 `json:"start_time"`
	EndTime       string                 `json:"end_time"`
	Address       string                 `json:"address"`
	UserID        int64                  `json:"user_id"`
	Nickname      string                 `json:"nickname"`
	Avatar        string                 `json:"avatar"`
	FormData      map[string]interface{} `json:"form_data"`
	Status        int8                   `json:"status"`
	StatusText    string                 `json:"status_text"`
	AuditRemark   string                 `json:"audit_remark"`
	AuditorName   string                 `json:"auditor_name"`
	AuditTime     string                 `json:"audit_time"`
	AppointTime   string                 `json:"appoint_time"`
	CancelTime    string                 `json:"cancel_time"`
	CreateTime    string                 `json:"create_time"`
}

type SignupDetail struct {
	SignupListItem
	FormConfig []model.FormField `json:"form_config"`
	Activity   *ActivityListItem `json:"activity"`
}

type ExportQuery struct {
	ActivityID int64 `json:"activity_id" form:"activity_id"`
	Status     int   `json:"status" form:"status"`
}
