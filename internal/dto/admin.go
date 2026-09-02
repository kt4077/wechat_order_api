package dto

// AdminLoginReq 管理端登录请求
type AdminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AdminLoginResp 管理端登录响应
type AdminLoginResp struct {
	Token    string     `json:"token"`
	ExpireAt int64      `json:"expire_at"`
	Profile  *AdminInfo `json:"profile"`
}

// AdminInfo 管理员信息
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

// AdminSaveReq 管理员新增/编辑请求
type AdminSaveReq struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Nickname    string   `json:"nickname"`
	Role        int8     `json:"role"`
	Status      int8     `json:"status"`
	Permissions []string `json:"permissions"`
}

// AdminQuery 管理员查询参数
type AdminQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// DashboardResp 后台数据总览
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

// TrendItem 趋势图数据项
type TrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// NameValueItem 名称-数值数据项
type NameValueItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// HotActivityItem 热门活动
type HotActivityItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	SignedCount int    `json:"signed_count"`
	Quota       int    `json:"quota"`
	Status      int8   `json:"status"`
}

// ConfigItem 系统配置项
type ConfigItem struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Remark      string `json:"remark"`
}

// MessageQuery 消息查询参数
type MessageQuery struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
	Type     int `json:"type" form:"type"`
}

// MessageItem 消息列表项
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
