package controller

import (
	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// Login 微信小程序登录（code 换取 openid，自动注册）
// @Summary 小程序登录
// @Tags 公共
func Login(c *gin.Context) {
	req, ok := bindJSON[dto.LoginReq](c)
	if !ok {
		return
	}
	resp, err := service.Login(req, clientIP(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, resp)
}

// GetConfig 获取小程序端系统配置（关于我们、隐私政策、使用帮助等）
func GetConfig(c *gin.Context) {
	data, err := service.GetConfigMap()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

// UploadImage 上传图片，返回可访问 URL
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, errcode.ErrParams.WithMsg("请选择要上传的图片"))
		return
	}
	url, err := uploadImageFile(file)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"url": url})
}

// Health 健康检查，用于容器探针与部署校验
func Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok", "service": "activity-api"})
}
