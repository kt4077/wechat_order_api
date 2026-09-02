// Package controller 负责参数接收、参数校验与结果返回，业务逻辑统一委托给 service 层。
package controller

import (
	"strconv"

	"activity/internal/middleware"
	"activity/internal/model"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// bindJSON 绑定并校验 JSON 请求体
func bindJSON[T any](c *gin.Context) (*T, bool) {
	req := new(T)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, errcode.ErrParams.WithMsg("请求参数格式错误"))
		return nil, false
	}
	return req, true
}

// bindQuery 绑定 URL 查询参数
func bindQuery[T any](c *gin.Context) (*T, bool) {
	req := new(T)
	if err := c.ShouldBindQuery(req); err != nil {
		response.Fail(c, errcode.ErrParams)
		return nil, false
	}
	return req, true
}

// pathID 解析路径中的主键参数
func pathID(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errcode.ErrParams.WithMsg("参数错误"))
		return 0, false
	}
	return id, true
}

// currentUser 获取当前登录用户 ID
func currentUser(c *gin.Context) int64 { return middleware.GetUserID(c) }

// isPlatformAdmin 实时判断当前小程序登录用户是否为平台超级管理员
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

// hasMerchantPermission 判断是否拥有主办方（活动发布）权限
func hasMerchantPermission(c *gin.Context) (int64, bool) {
	userID := currentUser(c)
	if userID == 0 {
		response.Fail(c, errcode.ErrUnauth)
		return 0, false
	}
	// 统一以数据库实时状态鉴权：
	// 令牌中的角色为登录时快照，入驻审核通过或权限被回收后需立即生效，因此不依赖令牌角色。
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

// clientIP 获取客户端 IP
func clientIP(c *gin.Context) string { return middleware.GetClientIP(c) }
