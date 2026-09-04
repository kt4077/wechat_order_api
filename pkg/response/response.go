package response

import (
	"net/http"

	"activity/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     "success",
		Data:    data,
		Success: true,
	})
}

func SuccessMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     msg,
		Data:    data,
		Success: true,
	})
}

func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Msg:     "success",
		Data:    PageData{List: list, Total: total, Page: page, PageSize: pageSize},
		Success: true,
	})
}

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

func FailMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil, Success: false})
}

func Abort(c *gin.Context, e *errcode.Error) {
	c.JSON(http.StatusOK, Response{Code: e.Code, Msg: e.Msg, Data: nil, Success: false})
	c.Abort()
}
