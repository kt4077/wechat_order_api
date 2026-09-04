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

	"encoding/json"

	"gorm.io/gorm"
)

var defaultConfigs = []dto.ConfigItem{
	{ConfigKey: "site_name", ConfigValue: "活动报名工具", Remark: "平台名称"},
	{ConfigKey: "icp_no", ConfigValue: "", Remark: "ICP 备案号"},
	{ConfigKey: "service_phone", ConfigValue: "", Remark: "客服电话"},
	{ConfigKey: "about_us", ConfigValue: "活动报名工具是一款面向个人、社团、培训机构与企业内训的轻量化报名小程序，提供活动发布、在线报名、审核统计、Excel 导出等一站式能力。", Remark: "关于我们"},
	{ConfigKey: "privacy_policy", ConfigValue: "我们严格遵循《个人信息保护法》与微信小程序平台规范，仅收集报名所必需的字段信息；手机号、身份证等隐私数据加密存储并做脱敏展示，不会向第三方披露。您可随时在「我的」中查看或删除个人报名数据。", Remark: "隐私政策"},
	{ConfigKey: "user_agreement", ConfigValue: "欢迎使用活动报名工具。使用本服务即表示您同意：如实填写报名信息、遵守平台活动发布规范、不发布违法违规内容。违规内容平台有权下架并追究责任。", Remark: "用户协议"},
	{ConfigKey: "help_doc", ConfigValue: "1. 报名流程：选择活动 → 填写表单 → 提交 → 等待审核 → 查看结果。\n2. 撤销报名：待审核状态下可在「我的报名」中撤销，名额自动释放。\n3. 申请入驻：「我的」→ 申请入驻 → 提交资料 → 平台审核 → 开通发布权限。", Remark: "使用帮助"},
}

func EnsureSuperAdmin() {
	cfg := config.Get().SuperAdmin
	if cfg.Username == "" {
		return
	}
	var count int64
	_ = model.DB.Model(&model.Admin{}).Count(&count).Error
	if count > 0 {
		return
	}
	hash, err := crypto.HashPassword(cfg.Password)
	if err != nil {
		logger.Warnf("初始化超级管理员失败：%v", err)
		return
	}
	nickname := cfg.Nickname
	if nickname == "" {
		nickname = "超级管理员"
	}
	if err := model.DB.Create(&model.Admin{
		Username: cfg.Username,
		Password: hash,
		Nickname: nickname,
		Role:     model.AdminRoleSuper,
		Status:   1,
	}).Error; err != nil {
		logger.Warnf("初始化超级管理员失败：%v", err)
		return
	}
	logger.Infof("已初始化超级管理员账号：%s / %s，请登录后立即修改密码", cfg.Username, cfg.Password)
}

func EnsureDefaultConfig() {
	for _, item := range defaultConfigs {
		var count int64
		_ = model.DB.Model(&model.SysConfig{}).Where("config_key = ?", item.ConfigKey).Count(&count).Error
		if count == 0 {
			_ = model.DB.Create(&model.SysConfig{
				ConfigKey:   item.ConfigKey,
				ConfigValue: item.ConfigValue,
				Remark:      item.Remark,
			}).Error
		}
	}
}

func AdminLogin(req *dto.AdminLoginReq, clientIP string) (*dto.AdminLoginResp, error) {
	req.Username = validate.Trim(req.Username)
	if req.Username == "" || req.Password == "" {
		return nil, errcode.ErrParams.WithMsg("请输入账号和密码")
	}
	admin := &model.Admin{}
	if err := model.DB.Where("username = ?", req.Username).First(admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrBusiness.WithMsg("账号或密码错误")
		}
		return nil, errcode.ErrSystem
	}
	if !crypto.CheckPassword(req.Password, admin.Password) {
		return nil, errcode.ErrBusiness.WithMsg("账号或密码错误")
	}
	if admin.Status != 1 {
		return nil, errcode.ErrForbid.WithMsg("账号已被禁用")
	}
	now := time.Now()
	if err := model.DB.Model(&model.Admin{}).Where("id = ?", admin.ID).
		Updates(map[string]interface{}{"last_login_time": now, "last_login_ip": clientIP}).Error; err != nil {
		logger.Warnf("更新管理员登录信息失败：%v", err)
	}
	token, expireAt, err := jwtx.GenerateAdminToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		return nil, errcode.ErrSystem.WithMsg("生成令牌失败")
	}
	return &dto.AdminLoginResp{
		Token:    token,
		ExpireAt: expireAt,
		Profile:  BuildAdminInfo(admin),
	}, nil
}

func GetAdminInfo(id int64) (*dto.AdminInfo, error) {
	admin := &model.Admin{}
	if err := model.DB.Where("id = ?", id).First(admin).Error; err != nil {
		return nil, errcode.ErrNotFound.WithMsg("账号不存在")
	}
	return BuildAdminInfo(admin), nil
}

func BuildAdminInfo(a *model.Admin) *dto.AdminInfo {
	perms := make([]string, 0)
	if a.Permissions != "" {
		_ = json.Unmarshal([]byte(a.Permissions), &perms)
	}
	roleText := "子管理员"
	if a.Role == model.AdminRoleSuper {
		roleText = "超级管理员"
	}
	return &dto.AdminInfo{
		ID:            a.ID,
		Username:      a.Username,
		Nickname:      a.Nickname,
		Avatar:        a.Avatar,
		Role:          a.Role,
		RoleText:      roleText,
		Status:        a.Status,
		Permissions:   perms,
		LastLoginTime: model.FmtTime(a.LastLoginTime),
	}
}

func ChangeAdminPassword(adminID int64, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errcode.ErrParams.WithMsg("新密码长度不能少于 6 位")
	}
	admin := &model.Admin{}
	if err := model.DB.Where("id = ?", adminID).First(admin).Error; err != nil {
		return errcode.ErrNotFound.WithMsg("账号不存在")
	}
	if !crypto.CheckPassword(oldPassword, admin.Password) {
		return errcode.ErrBusiness.WithMsg("原密码不正确")
	}
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errcode.ErrSystem.WithMsg("密码加密失败")
	}
	if err := model.DB.Model(&model.Admin{}).Where("id = ?", adminID).Update("password", hash).Error; err != nil {
		return errcode.ErrSystem.WithMsg("修改失败")
	}
	return nil
}

func AdminList(q *dto.AdminQuery) ([]dto.AdminInfo, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.Admin{})
	if q.Keyword != "" {
		tx = tx.Where("username LIKE ? OR nickname LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.Admin, 0)
	if err := tx.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.AdminInfo, 0, len(list))
	for i := range list {
		items = append(items, *BuildAdminInfo(&list[i]))
	}
	return items, total, nil
}

func SaveAdmin(operatorRole int8, req *dto.AdminSaveReq) (int64, error) {
	req.Username = validate.Trim(req.Username)
	req.Nickname = validate.Trim(req.Nickname)
	if req.Username == "" {
		return 0, errcode.ErrParams.WithMsg("请输入登录账号")
	}
	if len(req.Username) < 3 {
		return 0, errcode.ErrParams.WithMsg("账号长度不能少于 3 位")
	}
	if req.Role == 0 {
		req.Role = model.AdminRoleSub
	}
	if operatorRole != model.AdminRoleSuper {
		return 0, errcode.ErrForbid.WithMsg("仅超级管理员可管理账号")
	}
	permJSON, _ := json.Marshal(req.Permissions)

	if req.ID > 0 {
		admin := &model.Admin{}
		if err := model.DB.Where("id = ?", req.ID).First(admin).Error; err != nil {
			return 0, errcode.ErrNotFound.WithMsg("账号不存在")
		}
		if admin.Role == model.AdminRoleSuper && req.Role != model.AdminRoleSuper {
			return 0, errcode.ErrBusiness.WithMsg("不可降级超级管理员")
		}
		var dup int64
		_ = model.DB.Model(&model.Admin{}).Where("username = ? AND id <> ?", req.Username, req.ID).Count(&dup).Error
		if dup > 0 {
			return 0, errcode.ErrBusiness.WithMsg("登录账号已存在")
		}
		updates := map[string]interface{}{
			"username":    req.Username,
			"nickname":    req.Nickname,
			"role":        req.Role,
			"status":      req.Status,
			"permissions": string(permJSON),
		}
		if req.Status == 0 {
			updates["status"] = 1
		}
		if req.Password != "" {
			if len(req.Password) < 6 {
				return 0, errcode.ErrParams.WithMsg("密码长度不能少于 6 位")
			}
			hash, err := crypto.HashPassword(req.Password)
			if err != nil {
				return 0, errcode.ErrSystem.WithMsg("密码加密失败")
			}
			updates["password"] = hash
		}
		if err := model.DB.Model(&model.Admin{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			return 0, errcode.ErrSystem.WithMsg("保存失败")
		}
		return req.ID, nil
	}

	if req.Password == "" || len(req.Password) < 6 {
		return 0, errcode.ErrParams.WithMsg("密码长度不能少于 6 位")
	}
	var dup int64
	_ = model.DB.Model(&model.Admin{}).Where("username = ?", req.Username).Count(&dup).Error
	if dup > 0 {
		return 0, errcode.ErrBusiness.WithMsg("登录账号已存在")
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return 0, errcode.ErrSystem.WithMsg("密码加密失败")
	}
	admin := &model.Admin{
		Username:    req.Username,
		Password:    hash,
		Nickname:    req.Nickname,
		Role:        req.Role,
		Status:      1,
		Permissions: string(permJSON),
	}
	if err := model.DB.Create(admin).Error; err != nil {
		return 0, errcode.ErrSystem.WithMsg("创建失败")
	}
	return admin.ID, nil
}

func DeleteAdmin(operatorRole int8, id int64) error {
	if operatorRole != model.AdminRoleSuper {
		return errcode.ErrForbid.WithMsg("仅超级管理员可删除账号")
	}
	admin := &model.Admin{}
	if err := model.DB.Where("id = ?", id).First(admin).Error; err != nil {
		return errcode.ErrNotFound.WithMsg("账号不存在")
	}
	if admin.Role == model.AdminRoleSuper {
		return errcode.ErrBusiness.WithMsg("超级管理员不可删除")
	}
	if err := model.DB.Delete(admin).Error; err != nil {
		return errcode.ErrSystem.WithMsg("删除失败")
	}
	return nil
}

func GetConfigMap() (map[string]string, error) {
	list := make([]model.SysConfig, 0)
	if err := model.DB.Find(&list).Error; err != nil {
		return nil, errcode.ErrSystem
	}
	result := map[string]string{}
	for _, item := range list {
		result[item.ConfigKey] = item.ConfigValue
	}
	for _, item := range defaultConfigs {
		if _, ok := result[item.ConfigKey]; !ok {
			result[item.ConfigKey] = item.ConfigValue
		}
	}
	return result, nil
}

func SaveConfig(items []dto.ConfigItem) error {
	for _, item := range items {
		if item.ConfigKey == "" {
			continue
		}
		cfg := &model.SysConfig{}
		err := model.DB.Where("config_key = ?", item.ConfigKey).First(cfg).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = model.DB.Create(&model.SysConfig{ConfigKey: item.ConfigKey, ConfigValue: item.ConfigValue, Remark: item.Remark}).Error
			continue
		}
		_ = model.DB.Model(&model.SysConfig{}).Where("config_key = ?", item.ConfigKey).
			Update("config_value", item.ConfigValue).Error
	}
	return nil
}

func AdminUserList(q *dto.UserQuery) ([]dto.AdminUserItem, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.User{})
	if q.Keyword != "" {
		tx = tx.Where("nickname LIKE ?", "%"+q.Keyword+"%")
	}
	if q.Role != nil && *q.Role >= 0 {
		tx = tx.Where("role = ?", *q.Role)
	}
	if q.Status > 0 {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.User, 0)
	if err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.AdminUserItem, 0, len(list))
	for _, item := range list {
		var signupCount, activityCount int64
		_ = model.DB.Model(&model.Signup{}).Where("user_id = ?", item.ID).Count(&signupCount).Error
		_ = model.DB.Model(&model.Activity{}).Where("creator_id = ?", item.ID).Count(&activityCount).Error
		items = append(items, dto.AdminUserItem{
			ID:            item.ID,
			Nickname:      item.Nickname,
			Avatar:        item.Avatar,
			Phone:         crypto.Decrypt(item.Phone),
			Role:          item.Role,
			RoleText:      RoleText(item.Role),
			Status:        item.Status,
			MerchantFlag:  item.MerchantFlag,
			SignupCount:   signupCount,
			ActivityCount: activityCount,
			CreateTime:    model.FmtTimeValue(item.CreateTime),
		})
	}
	return items, total, nil
}

func ChangeUserStatus(adminID int64, adminName string, req *dto.ChangeStatusReq) error {
	if req.Status != model.UserStatusNormal && req.Status != model.UserStatusDisable {
		return errcode.ErrParams.WithMsg("状态不合法")
	}
	user, err := getUserByID(req.ID)
	if err != nil {
		return err
	}
	if user.Role == model.RoleSuperAdmin {
		return errcode.ErrBusiness.WithMsg("不可操作系统内置管理员")
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	action := "启用用户"
	tip := "您的账号已恢复正常，可继续使用报名服务。"
	if req.Status == model.UserStatusDisable {
		action = "禁用用户"
		tip = "您的账号因违规已被平台禁用，如有疑问请联系客服。"
	}
	_ = SendMessage(req.ID, model.MsgTypeSystem, "账号状态变更", tip, 0)
	RecordLog(adminID, adminName, "用户管理", action, "用户ID："+itoa(req.ID)+" "+validate.Trim(req.Remark), "")
	return nil
}

func RecordLog(adminID int64, adminName, module, action, detail, ip string) {
	log := &model.OperationLog{
		AdminID:   adminID,
		AdminName: adminName,
		Module:    module,
		Action:    action,
		Detail:    detail,
		IP:        ip,
	}
	if err := model.DB.Create(log).Error; err != nil {
		logger.Warnf("写入操作日志失败：%v", err)
	}
}

func OperationLogList(q *dto.AdminQuery) ([]model.OperationLog, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.OperationLog{})
	if q.Keyword != "" {
		tx = tx.Where("admin_name LIKE ? OR detail LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.OperationLog, 0)
	if err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	return list, total, nil
}
