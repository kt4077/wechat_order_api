package controller

import (
	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

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

func GetConfig(c *gin.Context) {
	data, err := service.GetConfigMap()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}

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

func Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok", "service": "activity-api"})
}
