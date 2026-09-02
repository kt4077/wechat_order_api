package controller

import (
	"strconv"

	"activity/internal/dto"
	"activity/internal/middleware"
	"activity/internal/model"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminLogin 管理端登录
func AdminLogin(c *gin.Context) {
	req, ok := bindJSON[dto.AdminLoginReq](c)
	if !ok {
		return
	}
	resp, err := service.AdminLogin(req, clientIP(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(resp.Profile.ID, resp.Profile.Nickname, "系统", "登录", "管理端登录成功", clientIP(c))
	response.Success(c, resp)
}

// AdminProfile 当前管理员信息
func AdminProfile(c *gin.Context) {
	data, err := service.GetAdminInfo(middleware.GetAdminID(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// ChangeAdminPassword 修改当前管理员密码
func ChangeAdminPassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrParams)
		return
	}
	if err := service.ChangeAdminPassword(middleware.GetAdminID(c), req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "密码修改成功", nil)
}

// Dashboard 后台数据总览
func Dashboard(c *gin.Context) {
	data, err := service.Dashboard()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// AdminList 管理员列表
func AdminList(c *gin.Context) {
	q, ok := bindQuery[dto.AdminQuery](c)
	if !ok {
		return
	}
	list, total, err := service.AdminList(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// SaveAdmin 新增 / 编辑管理员
func SaveAdmin(c *gin.Context) {
	req, ok := bindJSON[dto.AdminSaveReq](c)
	if !ok {
		return
	}
	id, err := service.SaveAdmin(middleware.GetAdminRole(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "权限管理", "保存账号",
		"账号ID："+strconv.FormatInt(id, 10)+" "+req.Username, clientIP(c))
	response.SuccessMsg(c, "保存成功", gin.H{"id": id})
}

// DeleteAdmin 删除管理员
func DeleteAdmin(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteAdmin(middleware.GetAdminRole(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "权限管理", "删除账号",
		"账号ID："+strconv.FormatInt(id, 10), clientIP(c))
	response.SuccessMsg(c, "删除成功", nil)
}

// GetSysConfig 获取系统配置列表
func GetSysConfig(c *gin.Context) {
	data, err := service.GetConfigMap()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// SaveSysConfig 保存系统配置
func SaveSysConfig(c *gin.Context) {
	var req struct {
		Items []dto.ConfigItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Items = nil
	}
	if len(req.Items) == 0 {
		// 兼容键值对对象提交方式
		kv := map[string]string{}
		if err := c.ShouldBindJSON(&kv); err == nil {
			for k, v := range kv {
				req.Items = append(req.Items, dto.ConfigItem{ConfigKey: k, ConfigValue: v})
			}
		}
	}
	if err := service.SaveConfig(req.Items); err != nil {
		response.Fail(c, err)
		return
	}
	service.RecordLog(middleware.GetAdminID(c), middleware.GetAdminName(c), "系统配置", "保存配置", "配置项数量："+strconv.Itoa(len(req.Items)), clientIP(c))
	response.SuccessMsg(c, "配置已保存", nil)
}

// OperationLogList 操作日志列表
func OperationLogList(c *gin.Context) {
	q, ok := bindQuery[dto.AdminQuery](c)
	if !ok {
		return
	}
	list, total, err := service.OperationLogList(q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// AdminUpload 管理端图片上传
func AdminUpload(c *gin.Context) {
	UploadImage(c)
}

// isSuperAdmin 判断是否超级管理员
func isSuperAdmin(c *gin.Context) bool {
	return middleware.GetAdminRole(c) == model.AdminRoleSuper
}
