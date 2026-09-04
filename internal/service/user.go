package service

import (
	"errors"
	"time"

	"activity/config"
	"activity/internal/dto"
	"activity/internal/model"
	"activity/pkg/crypto"
	"activity/pkg/errcode"
	"activity/pkg/jwtx"
	"activity/pkg/logger"
	"activity/pkg/validate"
	"activity/pkg/wechat"

	"gorm.io/gorm"
)

func Login(req *dto.LoginReq, clientIP string) (*dto.LoginResp, error) {
	if req.Code == "" {
		return nil, errcode.ErrParams.WithMsg("缺少登录凭证 code")
	}
	if req.Gender < 1 || req.Gender > 3 {
		req.Gender = model.GenderUnknown
	}
	openID := ""
	var unionID string
	if config.Get().Wechat.AppID == "" {
		openID = "dev_" + req.Code
	} else {
		session, err := wechat.Code2Session(req.Code)
		if err != nil {
			return nil, errcode.ErrBusiness.WithMsg(err.Error())
		}
		openID, unionID = session.OpenID, session.UnionID
	}
	if openID == "" {
		return nil, errcode.ErrBusiness.WithMsg("获取微信身份失败")
	}

	user := &model.User{}
	err := model.DB.Where("openid = ?", openID).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &model.User{
			Openid:   openID,
			Unionid:  unionID,
			Nickname: req.Nickname,
			Avatar:   req.Avatar,
			Gender:   req.Gender,
			Role:     model.RoleNormal,
			Status:   model.UserStatusNormal,
		}
		if user.Nickname == "" {
			user.Nickname = "微信用户" + openID[len(openID)-4:]
		}
		if err := model.DB.Create(user).Error; err != nil {
			return nil, errcode.ErrSystem.WithMsg("初始化用户信息失败")
		}
	} else if err != nil {
		return nil, errcode.ErrSystem
	} else {
		updates := map[string]interface{}{"last_login_time": time.Now()}
		if req.Nickname != "" && req.Nickname != user.Nickname {
			updates["nickname"] = req.Nickname
		}
		if req.Avatar != "" && req.Avatar != user.Avatar {
			updates["avatar"] = req.Avatar
		}
		if req.Gender > 0 {
			updates["gender"] = req.Gender
		}
		if err := model.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			logger.Warnf("更新用户登录信息失败：%v", err)
		}
	}

	if user.Status == model.UserStatusDisable {
		return nil, errcode.ErrForbid.WithMsg("账号已被禁用，请联系平台管理员")
	}

	token, expireAt, err := jwtx.GenerateUserToken(user.ID, user.Role)
	if err != nil {
		return nil, errcode.ErrSystem.WithMsg("生成登录令牌失败")
	}
	_ = clientIP
	info := BuildUserInfo(user, true)
	return &dto.LoginResp{Token: token, ExpireAt: expireAt, UserInfo: info, NeedPhone: user.Phone == ""}, nil
}

func GetProfile(userID int64) (*dto.UserInfo, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}
	return BuildUserInfo(user, true), nil
}

func UpdateProfile(userID int64, req *dto.UpdateProfileReq) (*dto.UserInfo, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = validate.Trim(req.Nickname)
	}
	if req.Avatar != "" {
		updates["avatar"] = validate.Trim(req.Avatar)
	}
	if req.Gender >= 1 && req.Gender <= 3 {
		updates["gender"] = req.Gender
	}
	if req.RealName != "" {
		updates["real_name"] = validate.Trim(req.RealName)
	}
	if req.Phone != "" {
		phone := validate.Trim(req.Phone)
		if !validate.IsPhone(phone) {
			return nil, errcode.ErrParams.WithMsg("手机号格式不正确")
		}
		cipher, err := crypto.Encrypt(phone)
		if err != nil {
			return nil, errcode.ErrSystem.WithMsg("手机号加密失败")
		}
		updates["phone"] = cipher
	}
	if len(updates) == 0 {
		return BuildUserInfo(user, true), nil
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, errcode.ErrSystem.WithMsg("更新资料失败")
	}
	user, err = getUserByID(userID)
	if err != nil {
		return nil, err
	}
	return BuildUserInfo(user, true), nil
}

func DeleteMyData(userID int64) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.Signup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errcode.ErrSystem.WithMsg("删除失败，请稍后重试")
	}
	return nil
}

func BuildUserInfo(u *model.User, self bool) *dto.UserInfo {
	phone := crypto.Decrypt(u.Phone)
	displayPhone := phone
	if !self {
		displayPhone = crypto.MaskPhone(phone)
	}
	return &dto.UserInfo{
		ID:           u.ID,
		Openid:       u.Openid,
		Nickname:     u.Nickname,
		Avatar:       u.Avatar,
		Phone:        displayPhone,
		RealName:     u.RealName,
		Gender:       u.Gender,
		Role:         u.Role,
		RoleText:     RoleText(u.Role),
		Status:       u.Status,
		MerchantFlag: u.MerchantFlag,
		CreateTime:   model.FmtTimeValue(u.CreateTime),
	}
}

func RoleText(role int8) string {
	switch role {
	case model.RoleMerchant:
		return "入驻管理员"
	case model.RoleSuperAdmin:
		return "超级管理员"
	default:
		return "普通用户"
	}
}

func getUserByID(id int64) (*model.User, error) {
	u := &model.User{}
	if err := model.DB.Where("id = ?", id).First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound.WithMsg("用户不存在")
		}
		return nil, errcode.ErrSystem
	}
	return u, nil
}
