package dto

type LoginReq struct {
	Code      string `json:"code"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	InviterID int64  `json:"inviter_id"`
}

type LoginResp struct {
	Token     string    `json:"token"`
	ExpireAt  int64     `json:"expire_at"`
	UserInfo  *UserInfo `json:"user_info"`
	NeedPhone bool      `json:"need_phone"`
}

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

type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
}

type UserQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Role     *int   `json:"role" form:"role"`
	Status   int    `json:"status" form:"status"`
}

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

type ChangeStatusReq struct {
	ID     int64  `json:"id"`
	Status int8   `json:"status"`
	Remark string `json:"remark"`
}
