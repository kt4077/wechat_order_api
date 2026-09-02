// Package errcode 定义全局统一错误码体系，区分参数错误、权限错误、业务错误与系统错误。
package errcode

// Error 业务错误类型
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string { return e.Msg }

// 错误码常量
const (
	CodeSuccess      = 0    // 成功
	CodeInvalidParam = 1001 // 参数错误
	CodeUnauthorized = 1002 // 未登录 / 登录已过期
	CodeForbidden    = 1003 // 权限不足
	CodeNotFound     = 1004 // 数据不存在
	CodeTooMany      = 1005 // 请求过于频繁
	CodeBusiness     = 2001 // 业务错误
	CodeSystem       = 5001 // 系统错误
)

// 预定义错误
var (
	OK           = &Error{Code: CodeSuccess, Msg: "success"}
	ErrParams    = &Error{Code: CodeInvalidParam, Msg: "参数错误"}
	ErrUnauth    = &Error{Code: CodeUnauthorized, Msg: "登录已过期，请重新登录"}
	ErrForbid    = &Error{Code: CodeForbidden, Msg: "权限不足"}
	ErrNotFound  = &Error{Code: CodeNotFound, Msg: "数据不存在"}
	ErrTooMany   = &Error{Code: CodeTooMany, Msg: "操作过于频繁，请稍后再试"}
	ErrBusiness  = &Error{Code: CodeBusiness, Msg: "业务处理失败"}
	ErrSystem    = &Error{Code: CodeSystem, Msg: "系统繁忙，请稍后再试"}
	ErrNoQuota   = &Error{Code: CodeBusiness, Msg: "报名名额已满"}
	ErrSigned    = &Error{Code: CodeBusiness, Msg: "您已报名该活动，请勿重复提交"}
	ErrNotInTime = &Error{Code: CodeBusiness, Msg: "当前不在报名时间范围内"}
	ErrEnded     = &Error{Code: CodeBusiness, Msg: "活动已结束，报名通道已关闭"}
	ErrPending   = &Error{Code: CodeBusiness, Msg: "已有审核中的申请，请耐心等待"}
)

// New 创建一个自定义业务错误
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// WithMsg 基于已有错误码生成新的错误提示，保持错误码语义不变
func (e *Error) WithMsg(msg string) *Error {
	return &Error{Code: e.Code, Msg: msg}
}

// Is 判断两个错误是否同一错误码
func Is(err error, target *Error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Code == target.Code
}
