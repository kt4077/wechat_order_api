package dto

// LoginReq 小程序登录请求
type LoginReq struct {
	Code      string `json:"code"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	InviterID int64  `json:"inviter_id"`
}

// LoginResp 小程序登录响应
type LoginResp struct {
	Token     string    `json:"token"`
	ExpireAt  int64     `json:"expire_at"`
	UserInfo  *UserInfo `json:"user_info"`
	NeedPhone bool      `json:"need_phone"`
}

// UserInfo 用户基本信息（对外输出，隐私字段已脱敏）
type UserInfo struct {
	ID           int64  `json:"id"`
	Openid       string `json:"openid"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Phone        string `json:"phone"`
	RealName     string `json:"real_name"`
	Gender       int8   `json:"gender"`
	Role         int8   `json:"role"`
	RoleText     string `json:"role_text"`
	Status       int8   `json:"status"`
	MerchantFlag int8   `json:"merchant_flag"`
	CreateTime   string `json:"create_time"`
}

// UpdateProfileReq 更新用户资料请求
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
}

// UserQuery 用户查询参数
type UserQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	// Role 使用指针类型：nil 表示查询全部角色
	Role   *int `json:"role" form:"role"`
	Status int  `json:"status" form:"status"` // 0 全部，1 正常，2 禁用
}

// AdminUserItem 后台用户列表项
type AdminUserItem struct {
	ID            int64  `json:"id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Phone         string `json:"phone"`
	Role          int8   `json:"role"`
	RoleText      string `json:"role_text"`
	Status        int8   `json:"status"`
	MerchantFlag  int8   `json:"merchant_flag"`
	SignupCount   int64  `json:"signup_count"`
	ActivityCount int64  `json:"activity_count"`
	CreateTime    string `json:"create_time"`
}

// ChangeStatusReq 状态变更通用请求
type ChangeStatusReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}
