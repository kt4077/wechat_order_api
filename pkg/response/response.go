// Package response 提供全局统一接口响应结构，固定字段 code / msg / data / success。
package response

import (
	"net/http"

	"activity/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Response 统一响应体
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

// PageData 分页响应数据
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     "success",
		Data:    data,
		Success: true,
	})
}

// SuccessMsg 成功响应并携带自定义提示
func SuccessMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     msg,
		Data:    data,
		Success: true,
	})
}

// Page 分页成功响应
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     "success",
		Data:    PageData{List: list, Total: total, Page: page, PageSize: pageSize},
		Success: true,
	})
}

// Fail 失败响应，自动识别业务错误与系统错误
func Fail(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}
	if e, ok := err.(*errcode.Error); ok {
		c.JSON(http.StatusOK, Response{Code: e.Code, Msg: e.Msg, Data: nil, Success: false})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSystem,
		Msg:     err.Error(),
		Data:    nil,
		Success: false,
	})
}

// FailMsg 自定义文案失败响应
func FailMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil, Success: false})
}

// Abort 常用于中间件直接中断请求并返回错误
func Abort(c *gin.Context, e *errcode.Error) {
	c.JSON(http.StatusOK, Response{Code: e.Code, Msg: e.Msg, Data: nil, Success: false})
	c.Abort()
}
