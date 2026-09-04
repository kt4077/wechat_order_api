package dto

type MerchantApplyReq struct {
	Type                int8     `json:"type"`
	ContactName         string   `json:"contact_name"`
	ContactPhone        string   `json:"contact_phone"`
	Intro               string   `json:"intro"`
	QualificationImages []string `json:"qualification_images"`
}

type MerchantAuditReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}

type MerchantBatchAuditReq struct {
	IDs    []int64 `json:"ids"`
	Status int8    `json:"status"`
	Remark string  `json:"remark"`
}

type MerchantQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   *int   `json:"status" form:"status"`
	Type     int    `json:"type" form:"type"`
}

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

const (
	MerchantDisplayNone     int8 = 1
	MerchantDisplayPending  int8 = 2
	MerchantDisplayPassed   int8 = 3
	MerchantDisplayRejected int8 = 4
	MerchantDisplayRevoked  int8 = 5
)

type MerchantStatusResp struct {
	Status       int8   `json:"status"`
	StatusText   string `json:"status_text"`
	ApplyID      int64  `json:"apply_id"`
	AuditRemark  string `json:"audit_remark"`
	MerchantFlag int8   `json:"merchant_flag"`
	CanApply     bool   `json:"can_apply"`
	CreateTime   string `json:"create_time"`
}
