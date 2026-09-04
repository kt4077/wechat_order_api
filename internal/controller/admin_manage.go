package controller

import (
	"net/url"
	"strconv"

	"activity/internal/dto"
	"activity/internal/middleware"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

func AdminActivityList(c *gin.Context) {
	q, ok := bindQuery[dto.ActivityQuery](c)
	if !ok {
		return
	}
	q.Scope = "all"
	list, total, err := service.ListActivities(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func AdminActivityDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	data, err := service.ActivityDetail(id, 0)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

func AdminChangeActivityStatus(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.ActivityStatusReq](c)
	if !ok {
		return
	}
	req.ID = id
	if err := service.ChangeActivityStatus(middleware.GetAdminID(c), true, req); err != nil {
		response.Fail(c, err)
		return
	}
	action := "上架活动"
	if req.Status == 4 {
		action = "下架活动"
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "活动管理", action,
		"活动ID："+strconv.FormatInt(id, 10)+" "+req.Remark, clientIP(c))
	response.SuccessMsg(c, "操作成功", nil)
}

func AdminAuditActivity(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.ActivityAuditReq](c)
	if !ok {
		return
	}
	req.ID = id
	if err := service.AuditActivity(req); err != nil {
		response.Fail(c, err)
		return
	}
	action := "审核通过活动"
	if req.Status == 2 {
		action = "驳回活动"
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "活动管理", action,
		"活动ID："+strconv.FormatInt(id, 10)+" "+req.Remark, clientIP(c))
	response.SuccessMsg(c, "审核完成", nil)
}

func AdminDeleteActivity(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteActivity(middleware.GetAdminID(c), true, id); err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "活动管理", "删除活动",
		"活动ID："+strconv.FormatInt(id, 10), clientIP(c))
	response.SuccessMsg(c, "删除成功", nil)
}

func AdminSignupList(c *gin.Context) {
	q, ok := bindQuery[dto.SignupQuery](c)
	if !ok {
		return
	}
	list, total, err := service.ListSignups(q, true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func AdminSignupDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	data, err := service.SignupDetail(id, 0, true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

func AdminAuditSignup(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.SignupAuditReq](c)
	if !ok {
		return
	}
	req.ID = id
	if err := service.AuditSignup(middleware.GetAdminID(c), middleware.GetAdminName(c), true, req); err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "报名管理", "审核报名",
		"报名ID："+strconv.FormatInt(id, 10)+" 结果："+strconv.Itoa(int(req.Status)), clientIP(c))
	response.SuccessMsg(c, "审核完成", nil)
}

func AdminBatchAuditSignups(c *gin.Context) {
	req, ok := bindJSON[dto.SignupBatchAuditReq](c)
	if !ok {
		return
	}
	count, err := service.BatchAuditSignups(middleware.GetAdminID(c), middleware.GetAdminName(c), true, req)
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "报名管理", "批量审核报名",
		"数量："+strconv.Itoa(count), clientIP(c))
	if err != nil {
		response.FailMsg(c, errcode.CodeBusiness, err.Error()+"，其余记录已完成审核")
		return
	}
	response.SuccessMsg(c, "成功审核 "+strconv.Itoa(count)+" 条记录", gin.H{"count": count})
}

func AdminExportSignups(c *gin.Context) {
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
	data, fileName, err := service.ExportSignups(id, status, middleware.GetAdminID(c), true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "报名管理", "导出数据",
		"活动ID："+strconv.FormatInt(id, 10), clientIP(c))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func AdminUserList(c *gin.Context) {
	q, ok := bindQuery[dto.UserQuery](c)
	if !ok {
		return
	}
	list, total, err := service.AdminUserList(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func AdminChangeUserStatus(c *gin.Context) {
	req, ok := bindJSON[dto.ChangeStatusReq](c)
	if !ok {
		return
	}
	if err := service.ChangeUserStatus(middleware.GetAdminID(c), middleware.GetAdminName(c), req); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "操作成功", nil)
}

func AdminRevokeMerchant(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.RevokeMerchant(middleware.GetAdminID(c), middleware.GetAdminName(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "已收回该用户的入驻权限", nil)
}

func AdminApplyList(c *gin.Context) {
	q, ok := bindQuery[dto.MerchantQuery](c)
	if !ok {
		return
	}
	list, total, err := service.AdminApplyList(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func AdminAuditMerchant(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	req, ok := bindJSON[dto.MerchantAuditReq](c)
	if !ok {
		return
	}
	req.ID = id
	if err := service.AuditMerchant(middleware.GetAdminID(c), middleware.GetAdminName(c), req); err != nil {
		response.Fail(c, err)
		return
	}
	action := "通过入驻申请"
	if req.Status == 2 {
		action = "驳回入驻申请"
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "入驻审核", action,
		"申请ID："+strconv.FormatInt(id, 10), clientIP(c))
	response.SuccessMsg(c, "审核完成", nil)
}

func AdminBatchAuditMerchant(c *gin.Context) {
	req, ok := bindJSON[dto.MerchantBatchAuditReq](c)
	if !ok {
		return
	}
	count, err := service.BatchAuditMerchant(middleware.GetAdminID(c), middleware.GetAdminName(c), req)
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "入驻审核", "批量审核入驻申请",
		"数量："+strconv.Itoa(count), clientIP(c))
	if err != nil {
		response.FailMsg(c, errcode.CodeBusiness, err.Error()+"，其余记录已完成审核")
		return
	}
	response.SuccessMsg(c, "成功审核 "+strconv.Itoa(count)+" 条申请", gin.H{"count": count})
}
