package dto

// MerchantApplyReq 提交入驻申请请求
type MerchantApplyReq struct {
	Type                int8     `json:"type"`
	ContactName         string   `json:"contact_name"`
	ContactPhone        string   `json:"contact_phone"`
	Intro               string   `json:"intro"`
	QualificationImages []string `json:"qualification_images"`
}

// MerchantAuditReq 入驻审核请求
type MerchantAuditReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"` // 1 通过 2 驳回
	Remark string `json:"remark"`
}

// MerchantBatchAuditReq 入驻批量审核请求
type MerchantBatchAuditReq struct {
	IDs    []int64 `json:"ids"`
	Status int8    `json:"status"`
	Remark string  `json:"remark"`
}

// MerchantQuery 入驻申请查询参数
type MerchantQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	// Status 使用指针类型：nil 表示查询全部
	Status *int `json:"status" form:"status"`
	Type   int  `json:"type" form:"type"` // 0 全部
}

// MerchantApplyItem 入驻申请列表项
type MerchantApplyItem struct {
	ID                  int64    `json:"id"`
	UserID              int64    `json:"user_id"`
	Nickname            string   `json:"nickname"`
	Avatar              string   `json:"avatar"`
	Type                int8     `json:"type"`
	TypeText            string   `json:"type_text"`
	ContactName         string   `json:"contact_name"`
	ContactPhone        string   `json:"contact_phone"`
	Intro               string   `json:"intro"`
	QualificationImages []string `json:"qualification_images"`
	Status              int8     `json:"status"`
	StatusText          string   `json:"status_text"`
	AuditRemark         string   `json:"audit_remark"`
	AuditorName         string   `json:"auditor_name"`
	AuditTime           string   `json:"audit_time"`
	CreateTime          string   `json:"create_time"`
}

// 用户端入驻展示状态码（从 1 开始，避免使用 0 作为状态值）
const (
	MerchantDisplayNone     int8 = 1 // 未申请
	MerchantDisplayPending  int8 = 2 // 待审核
	MerchantDisplayPassed   int8 = 3 // 已通过
	MerchantDisplayRejected int8 = 4 // 已驳回
	MerchantDisplayRevoked  int8 = 5 // 权限已收回
)

// MerchantStatusResp 用户端入驻状态
type MerchantStatusResp struct {
	Status       int8   `json:"status"` // 1 未申请 2 待审核 3 通过 4 驳回 5 权限收回
	StatusText   string `json:"status_text"`
	ApplyID      int64  `json:"apply_id"`
	AuditRemark  string `json:"audit_remark"`
	MerchantFlag int8   `json:"merchant_flag"`
	CanApply     bool   `json:"can_apply"`
	CreateTime   string `json:"create_time"`
}
