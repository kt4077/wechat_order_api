package controller

import (
	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListActivities 活动列表（首页 / 我的活动）
func ListActivities(c *gin.Context) {
	q, ok := bindQuery[dto.ActivityQuery](c)
	if !ok {
		return
	}
	q.UserID = currentUser(c)
	list, total, err := service.ListActivities(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// ActivityDetail 活动详情
func ActivityDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	data, err := service.ActivityDetail(id, currentUser(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// CreateActivity 发布活动（需入驻权限）
func CreateActivity(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	req, ok := bindJSON[dto.ActivitySaveReq](c)
	if !ok {
		return
	}
	req.ID = 0
	id, err := service.SaveActivity(userID, isPlatformAdmin(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "活动创建成功", gin.H{"id": id})
}

// UpdateActivity 编辑活动（仅创建者或平台管理员）
func UpdateActivity(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.ActivitySaveReq](c)
	if !ok {
		return
	}
	req.ID = id
	newID, err := service.SaveActivity(userID, isPlatformAdmin(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "保存成功", gin.H{"id": newID})
}

// ChangeActivityStatus 上架 / 下架活动
func ChangeActivityStatus(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.ActivityStatusReq](c)
	if !ok {
		return
	}
	req.ID = id
	if err := service.ChangeActivityStatus(userID, isPlatformAdmin(c), req); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "操作成功", nil)
}

// DeleteActivity 删除活动（软删除）
func DeleteActivity(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteActivity(userID, isPlatformAdmin(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "删除成功", nil)
}

// CopyActivity 复刻活动
func CopyActivity(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	newID, err := service.CopyActivity(userID, isPlatformAdmin(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "复刻成功", gin.H{"id": newID})
}

// ActivityStat 单活动数据统计
func ActivityStat(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	data, err := service.GetActivityStat(id, currentUser(c), isPlatformAdmin(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}
