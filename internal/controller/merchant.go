package controller

import (
	"activity/internal/dto"
	"activity/internal/service"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// ApplyMerchant 提交入驻申请
func ApplyMerchant(c *gin.Context) {
	userID := currentUser(c)
	if userID == 0 {
		response.Fail(c, errcode.ErrUnauth)
		return
	}
	req, ok := bindJSON[dto.MerchantApplyReq](c)
	if !ok {
		return
	}
	if err := service.ApplyMerchant(userID, req, clientIP(c)); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessMsg(c, "申请已提交，请等待平台审核", nil)
}

// MyMerchantStatus 查询我的入驻状态
func MyMerchantStatus(c *gin.Context) {
	data, err := service.MyMerchantStatus(currentUser(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, data)
}
