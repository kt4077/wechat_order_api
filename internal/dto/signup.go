package dto

import "activity/internal/model"

// SignupSubmitReq 提交报名请求
type SignupSubmitReq struct {
	ActivityID int64                  `json:"activity_id"`
	FormData   map[string]interface{} `json:"form_data"`
	// AppointTime 报名者自行填写的预约到场时间（选填，格式 2006-01-02 15:04:05）
	AppointTime string `json:"appoint_time"`
}

// SignupAuditReq 报名审核请求
type SignupAuditReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"` // 1 通过 2 驳回
	Remark string `json:"remark"`
}

// SignupBatchAuditReq 报名批量审核请求
type SignupBatchAuditReq struct {
	IDs    []int64 `json:"ids"`
	Status int8    `json:"status"`
	Remark string  `json:"remark"`
}

// SignupQuery 报名记录查询参数
type SignupQuery struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	ActivityID int64  `json:"activity_id" form:"activity_id"`
	Keyword    string `json:"keyword" form:"keyword"`
	// Status 使用指针类型：nil 表示查询全部，避免零值 0 被误判为「待审核」
	Status *int   `json:"status" form:"status"`
	Scope  string `json:"scope" form:"scope"` // mine：我的报名
	UserID int64  `json:"-" form:"-"`
}

// SignupListItem 报名列表项
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

// SignupDetail 报名详情，附带表单配置用于回显
type SignupDetail struct {
	SignupListItem
	FormConfig []model.FormField `json:"form_config"`
	Activity   *ActivityListItem `json:"activity"`
}

// ExportQuery 导出参数
type ExportQuery struct {
	ActivityID int64 `json:"activity_id" form:"activity_id"`
	Status     int   `json:"status" form:"status"`
}
