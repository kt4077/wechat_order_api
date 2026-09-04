package controller

import (
	"strconv"

	"activity/internal/middleware"
	"activity/internal/model"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

func bindJSON[T any](c *gin.Context) (*T, bool) {
	req := new(T)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, errcode.ErrParams.WithMsg("请求参数格式错误"))
		return nil, false
	}
	return req, true
}

func bindQuery[T any](c *gin.Context) (*T, bool) {
	req := new(T)
	if err := c.ShouldBindQuery(req); err != nil {
		response.Fail(c, errcode.ErrParams)
		return nil, false
	}
	return req, true
}

func pathID(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errcode.ErrParams.WithMsg("参数错误"))
		return 0, false
	}
	return id, true
}

func currentUser(c *gin.Context) int64 { return middleware.GetUserID(c) }

func isPlatformAdmin(c *gin.Context) bool {
	userID := currentUser(c)
	if userID == 0 {
		return false
	}
	var role int8
	if err := model.DB.Model(&model.User{}).Select("role").Where("id = ?", userID).Scan(&role).Error; err != nil {
		return false
	}
	return role == model.RoleSuperAdmin
}

func hasMerchantPermission(c *gin.Context) (int64, bool) {
	userID := currentUser(c)
	if userID == 0 {
		response.Fail(c, errcode.ErrUnauth)
		return 0, false
	}
	user := &model.User{}
	if err := model.DB.Select("id", "merchant_flag", "role", "status").Where("id = ?", userID).First(user).Error; err != nil {
		response.Fail(c, errcode.ErrUnauth)
		return 0, false
	}
	if user.Status == model.UserStatusDisable {
		response.Fail(c, errcode.ErrForbid.WithMsg("账号已被禁用"))
		return 0, false
	}
	if user.Role != model.RoleSuperAdmin && (user.Role != model.RoleMerchant || user.MerchantFlag != 2) {
		response.Fail(c, errcode.ErrForbid.WithMsg("请先申请入驻并通过审核后再发布活动"))
		return 0, false
	}
	return userID, true
}

func clientIP(c *gin.Context) string { return middleware.GetClientIP(c) }
