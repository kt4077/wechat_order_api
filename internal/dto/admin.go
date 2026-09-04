package dto

type AdminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResp struct {
	Token    string     `json:"token"`
	ExpireAt int64      `json:"expire_at"`
	Profile  *AdminInfo `json:"profile"`
}

type AdminInfo struct {
	ID            int64    `json:"id"`
	Username      string   `json:"username"`
	Nickname      string   `json:"nickname"`
	Avatar        string   `json:"avatar"`
	Role          int8     `json:"role"`
	RoleText      string   `json:"role_text"`
	Status        int8     `json:"status"`
	Permissions   []string `json:"permissions"`
	LastLoginTime string   `json:"last_login_time"`
}

type AdminSaveReq struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Nickname    string   `json:"nickname"`
	Role        int8     `json:"role"`
	Status      int8     `json:"status"`
	Permissions []string `json:"permissions"`
}

type AdminQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
}

type DashboardResp struct {
	ActivityTotal   int64             `json:"activity_total"`
	ActivitySigning int64             `json:"activity_signing"`
	SignupTotal     int64             `json:"signup_total"`
	SignupPending   int64             `json:"signup_pending"`
	UserTotal       int64             `json:"user_total"`
	MerchantTotal   int64             `json:"merchant_total"`
	ApplyPending    int64             `json:"apply_pending"`
	TodaySignup     int64             `json:"today_signup"`
	Trend           []TrendItem       `json:"trend"`
	StatusDist      []NameValueItem   `json:"status_dist"`
	CategoryDist    []NameValueItem   `json:"category_dist"`
	HotActivities   []HotActivityItem `json:"hot_activities"`
}

type TrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type NameValueItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type HotActivityItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	SignedCount int    `json:"signed_count"`
	Quota       int    `json:"quota"`
	Status      int8   `json:"status"`
}

type ConfigItem struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Remark      string `json:"remark"`
}

type MessageQuery struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
	Type     int `json:"type" form:"type"`
}

type MessageItem struct {
	ID         int64  `json:"id"`
	Type       int8   `json:"type"`
	TypeText   string `json:"type_text"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	RelatedID  int64  `json:"related_id"`
	IsRead     int8   `json:"is_read"`
	CreateTime string `json:"create_time"`
}
