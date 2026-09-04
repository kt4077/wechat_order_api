package controller

import (
	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

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
