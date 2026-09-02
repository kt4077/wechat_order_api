package controller

import (
	"net/url"
	"strconv"

	"activity/internal/dto"
	"activity/internal/model"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// SubmitSignup 提交报名
func SubmitSignup(c *gin.Context) {
	userID := currentUser(c)
	if userID == 0 {
		response.Fail(c, errcode.ErrUnauth)
		return
	}
	req, ok := bindJSON[dto.SignupSubmitReq](c)
	if !ok {
		return
	}
	id, err := service.SubmitSignup(userID, req, clientIP(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "报名提交成功，请等待审核", gin.H{"id": id})
}

// ListSignups 报名记录列表（我的报名 / 活动报名管理）
func ListSignups(c *gin.Context) {
	q, ok := bindQuery[dto.SignupQuery](c)
	if !ok {
		return
	}
	q.UserID = currentUser(c)
	isAdmin := isPlatformAdmin(c)
	// 路由 /activities/:id/signups：活动报名管理，需按活动过滤并校验归属（仅活动创建者或平台管理员可查看）
	// 注意：/signups（我的报名）路由没有 :id 参数，因此先判断参数是否存在再解析。
	if v := c.Param("id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.Fail(c, errcode.ErrParams.WithMsg("参数错误"))
			return
		}
		q.ActivityID = id
		if !isAdmin {
			var a model.Activity
			if err := model.DB.Select("id", "creator_id").Where("id = ?", id).First(&a).Error; err != nil {
				response.Fail(c, errcode.ErrNotFound.WithMsg("活动不存在"))
				return
			}
			if a.CreatorID != q.UserID {
				response.Fail(c, errcode.ErrForbid.WithMsg("仅活动创建者可查看报名记录"))
				return
			}
		}
	}
	list, total, err := service.ListSignups(q, isAdmin)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// SignupDetail 报名详情
func SignupDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	data, err := service.SignupDetail(id, currentUser(c), isPlatformAdmin(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// CancelSignup 撤销报名
func CancelSignup(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.CancelSignup(currentUser(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "已撤销报名，名额已释放", nil)
}

// AuditSignup 审核报名（活动创建者或平台管理员）
func AuditSignup(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.SignupAuditReq](c)
	if !ok {
		return
	}
	req.ID = id
	profile, _ := service.GetProfile(userID)
	name := ""
	if profile != nil {
		name = profile.Nickname
	}
	if err := service.AuditSignup(userID, name, isPlatformAdmin(c), req); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "审核完成", nil)
}

// BatchAuditSignups 批量审核报名
func BatchAuditSignups(c *gin.Context) {
	userID, ok := hasMerchantPermission(c)
	if !ok {
		return
	}
	req, ok := bindJSON[dto.SignupBatchAuditReq](c)
	if !ok {
		return
	}
	profile, _ := service.GetProfile(userID)
	name := ""
	if profile != nil {
		name = profile.Nickname
	}
	count, err := service.BatchAuditSignups(userID, name, isPlatformAdmin(c), req)
	if err != nil {
		response.FailMsg(c, errcode.CodeBusiness, err.Error()+"，其余记录已完成审核")
		return
	}
	response.SuccessMsg(c, "成功审核 "+strconv.Itoa(count)+" 条记录", gin.H{"count": count})
}

// ExportSignups 导出活动报名数据为 Excel
func ExportSignups(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	status := -1
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = v
		}
	}
	data, fileName, err := service.ExportSignups(id, status, currentUser(c), isPlatformAdmin(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
