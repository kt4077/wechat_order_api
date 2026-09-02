package controller

import (
	"strconv"

	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetProfile 获取个人资料
func GetProfile(c *gin.Context) {
	data, err := service.GetProfile(currentUser(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// UpdateProfile 更新个人资料
func UpdateProfile(c *gin.Context) {
	req, ok := bindJSON[dto.UpdateProfileReq](c)
	if !ok {
		return
	}
	data, err := service.UpdateProfile(currentUser(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "保存成功", data)
}

// DeleteMyData 用户自主删除个人报名数据（软删除）
func DeleteMyData(c *gin.Context) {
	if err := service.DeleteMyData(currentUser(c)); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "已删除个人报名数据", nil)
}

// ListMessages 消息通知列表
func ListMessages(c *gin.Context) {
	q, ok := bindQuery[dto.MessageQuery](c)
	if !ok {
		return
	}
	list, total, err := service.ListMessages(currentUser(c), q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// ReadMessage 标记消息已读，支持 ?id=1 或请求体 {"id":1}
func ReadMessage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		var body struct {
			ID int64 `json:"id"`
		}
		if bErr := c.ShouldBindJSON(&body); bErr != nil || body.ID <= 0 {
			response.Fail(c, errcode.ErrParams)
			return
		}
		id = body.ID
	}
	if err := service.ReadMessage(currentUser(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "已标记已读", nil)
}

// ReadAllMessages 全部标记已读
func ReadAllMessages(c *gin.Context) {
	if err := service.ReadAllMessages(currentUser(c)); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "已全部标记已读", nil)
}

// UnreadCount 未读消息数量
func UnreadCount(c *gin.Context) {
	response.Success(c, gin.H{"count": service.UnreadCount(currentUser(c))})
}
