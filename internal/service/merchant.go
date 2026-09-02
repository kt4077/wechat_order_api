package service

import (
	"errors"

	"activity/config"
	"activity/internal/dto"
	"activity/internal/model"
	"activity/pkg/crypto"
	"activity/pkg/errcode"
	"activity/pkg/logger"
	"activity/pkg/validate"
	"activity/pkg/wechat"

	"encoding/json"

	"gorm.io/gorm"
)

// MerchantStatusText 用户端入驻「展示状态」文案（对应 dto.MerchantDisplayXxx）
func MerchantStatusText(status int8) string {
	switch status {
	case dto.MerchantDisplayNone:
		return "未申请"
	case dto.MerchantDisplayPending:
		return "审核中"
	case dto.MerchantDisplayPassed:
		return "审核通过"
	case dto.MerchantDisplayRejected:
		return "已驳回"
	case dto.MerchantDisplayRevoked:
		return "入驻权限已收回"
	default:
		return "未知"
	}
}

// ApplyStatusText 后台入驻申请「数据库状态」文案（对应 model.ApplyStatusXxx）
func ApplyStatusText(status int8) string {
	switch status {
	case model.ApplyStatusPending:
		return "待审核"
	case model.ApplyStatusApproved:
		return "审核通过"
	case model.ApplyStatusRejected:
		return "审核驳回"
	default:
		return "未知"
	}
}

// MerchantTypeText 入驻类型文案
func MerchantTypeText(t int8) string {
	if t == model.MerchantTypeOrg {
		return "机构"
	}
	return "个人"
}

// ApplyMerchant 提交入驻申请。
// 规则：同一用户同一时间仅允许一条待审核申请；驳回后可无限次重新提交并覆盖数据；入驻成功后不可重复申请。
func ApplyMerchant(userID int64, req *dto.MerchantApplyReq, clientIP string) error {
	req.ContactName = validate.Trim(req.ContactName)
	req.ContactPhone = validate.Trim(req.ContactPhone)
	if req.ContactName == "" {
		return errcode.ErrParams.WithMsg("请输入联系人姓名")
	}
	if !validate.IsPhone(req.ContactPhone) {
		return errcode.ErrParams.WithMsg("请输入正确的联系电话")
	}
	if req.Type != model.MerchantTypePerson && req.Type != model.MerchantTypeOrg {
		return errcode.ErrParams.WithMsg("请选择入驻类型")
	}
	if len([]rune(req.Intro)) > 500 {
		return errcode.ErrParams.WithMsg("申请说明不能超过 500 字")
	}
	if len(req.QualificationImages) > 9 {
		return errcode.ErrParams.WithMsg("资质图片最多上传 9 张")
	}

	user, err := getUserByID(userID)
	if err != nil {
		return err
	}
	if user.Status == model.UserStatusDisable {
		return errcode.ErrForbid.WithMsg("账号已被禁用，无法申请入驻")
	}
	if user.Role == model.RoleMerchant && user.MerchantFlag == 2 {
		return errcode.ErrBusiness.WithMsg("您已是入驻主办方，无需重复申请")
	}

	// 待审核申请唯一性校验
	var pendingCount int64
	if err := model.DB.Model(&model.MerchantApply{}).
		Where("user_id = ? AND status = ?", userID, model.ApplyStatusPending).Count(&pendingCount).Error; err != nil {
		return errcode.ErrSystem
	}
	if pendingCount > 0 {
		return errcode.ErrPending
	}

	images := req.QualificationImages
	if images == nil {
		images = []string{}
	}
	imageJSON, _ := json.Marshal(images)
	phone, err := crypto.Encrypt(req.ContactPhone)
	if err != nil {
		return errcode.ErrSystem.WithMsg("联系方式加密失败")
	}

	apply := &model.MerchantApply{
		UserID:              userID,
		Nickname:            user.Nickname,
		Avatar:              user.Avatar,
		Type:                req.Type,
		ContactName:         req.ContactName,
		ContactPhone:        phone,
		Intro:               validate.Trim(req.Intro),
		QualificationImages: string(imageJSON),
		Status:              model.ApplyStatusPending,
		ClientIP:            clientIP,
	}
	if err := model.DB.Create(apply).Error; err != nil {
		logger.Errorf("创建入驻申请失败：%v", err)
		return errcode.ErrSystem.WithMsg("提交申请失败，请稍后重试")
	}
	_ = SendMessage(userID, model.MsgTypeMerchant, "入驻申请已提交",
		"您的主办方入驻申请已提交，平台将在 1-3 个工作日内完成审核，请耐心等待。", apply.ID)
	return nil
}

// MyMerchantStatus 查询当前用户的入驻状态与最新申请信息
func MyMerchantStatus(userID int64) (*dto.MerchantStatusResp, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}
	resp := &dto.MerchantStatusResp{Status: dto.MerchantDisplayNone, StatusText: MerchantStatusText(dto.MerchantDisplayNone), MerchantFlag: user.MerchantFlag}
	apply := &model.MerchantApply{}
	err = model.DB.Where("user_id = ?", userID).Order("id DESC").First(apply).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		resp.CanApply = user.MerchantFlag == 1
		resp.StatusText = MerchantStatusText(dto.MerchantDisplayNone)
		return resp, nil
	}
	if err != nil {
		return nil, errcode.ErrSystem
	}
	resp.ApplyID = apply.ID
	resp.AuditRemark = apply.AuditRemark
	resp.CreateTime = model.FmtTimeValue(apply.CreateTime)

	switch apply.Status {
	case model.ApplyStatusPending:
		resp.Status = dto.MerchantDisplayPending
		resp.CanApply = false
	case model.ApplyStatusApproved:
		if user.MerchantFlag == 2 {
			resp.Status = dto.MerchantDisplayPassed
			resp.CanApply = false
		} else {
			// 平台已收回入驻权限
			resp.Status = dto.MerchantDisplayRevoked
			resp.CanApply = false
		}
	case model.ApplyStatusRejected:
		resp.Status = dto.MerchantDisplayRejected
		resp.CanApply = true
	}
	resp.StatusText = MerchantStatusText(resp.Status)
	return resp, nil
}

// AdminApplyList 后台入驻申请列表
func AdminApplyList(q *dto.MerchantQuery) ([]dto.MerchantApplyItem, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.MerchantApply{})
	if q.Keyword != "" {
		tx = tx.Where("(nickname LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?)",
			"%"+q.Keyword+"%", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if q.Status != nil && *q.Status >= 0 {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Type > 0 {
		tx = tx.Where("type = ?", q.Type)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.MerchantApply, 0)
	if err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.MerchantApplyItem, 0, len(list))
	for _, item := range list {
		items = append(items, *BuildMerchantApplyItem(&item))
	}
	return items, total, nil
}

// BuildMerchantApplyItem 组装入驻申请列表项，后台展示完整联系电话
func BuildMerchantApplyItem(a *model.MerchantApply) *dto.MerchantApplyItem {
	images := make([]string, 0)
	if a.QualificationImages != "" {
		_ = json.Unmarshal([]byte(a.QualificationImages), &images)
	}
	return &dto.MerchantApplyItem{
		ID:                  a.ID,
		UserID:              a.UserID,
		Nickname:            a.Nickname,
		Avatar:              a.Avatar,
		Type:                a.Type,
		TypeText:            MerchantTypeText(a.Type),
		ContactName:         a.ContactName,
		ContactPhone:        crypto.Decrypt(a.ContactPhone),
		Intro:               a.Intro,
		QualificationImages: images,
		Status:              a.Status,
		StatusText:          ApplyStatusText(a.Status),
		AuditRemark:         a.AuditRemark,
		AuditorName:         a.AuditorName,
		AuditTime:           model.FmtTime(a.AuditTime),
		CreateTime:          model.FmtTimeValue(a.CreateTime),
	}
}

// AuditMerchant 审核入驻申请：通过后自动开通主办方权限并推送通知。
// 入参 req.Status 为「审核动作」（1通过 / 2驳回），内部映射为存储状态。
func AuditMerchant(adminID int64, adminName string, req *dto.MerchantAuditReq) error {
	if req.Status != dto.AuditActionPass && req.Status != dto.AuditActionReject {
		return errcode.ErrParams.WithMsg("审核状态不合法")
	}
	if req.Status == dto.AuditActionReject && validate.Trim(req.Remark) == "" {
		return errcode.ErrParams.WithMsg("驳回时请填写驳回原因")
	}
	// 动作 → 存储状态
	targetStatus := model.ApplyStatusApproved
	if req.Status == dto.AuditActionReject {
		targetStatus = model.ApplyStatusRejected
	}
	var target *model.MerchantApply
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		apply := &model.MerchantApply{}
		if err := tx.Clauses(clauseForUpdate()).Where("id = ?", req.ID).First(apply).Error; err != nil {
			return errcode.ErrNotFound.WithMsg("申请记录不存在")
		}
		if apply.Status != model.ApplyStatusPending {
			return errcode.ErrBusiness.WithMsg("该申请已审核，请勿重复操作")
		}
		now := nowTime()
		if err := tx.Model(&model.MerchantApply{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"status":       targetStatus,
			"audit_remark": validate.Trim(req.Remark),
			"auditor_id":   adminID,
			"auditor_name": adminName,
			"audit_time":   now,
		}).Error; err != nil {
			return errcode.ErrSystem.WithMsg("审核失败")
		}
		if targetStatus == model.ApplyStatusApproved {
			// 开通主办方权限
			if err := tx.Model(&model.User{}).Where("id = ?", apply.UserID).
				Updates(map[string]interface{}{
					"role":          model.RoleMerchant,
					"merchant_flag": 2,
					"real_name":     apply.ContactName,
				}).Error; err != nil {
				return errcode.ErrSystem.WithMsg("权限开通失败")
			}
		}
		apply.Status = targetStatus
		apply.AuditRemark = req.Remark
		target = apply
		return nil
	})
	if err != nil {
		return err
	}
	notifyMerchantAuditResult(target, adminName)
	return nil
}

// BatchAuditMerchant 批量审核入驻申请
func BatchAuditMerchant(adminID int64, adminName string, req *dto.MerchantBatchAuditReq) (int, error) {
	if len(req.IDs) == 0 {
		return 0, errcode.ErrParams.WithMsg("请选择要审核的申请")
	}
	success := 0
	var firstErr error
	for _, id := range req.IDs {
		if err := AuditMerchant(adminID, adminName, &dto.MerchantAuditReq{
			ID:     id,
			Status: req.Status,
			Remark: req.Remark,
		}); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		success++
	}
	return success, firstErr
}

// RevokeMerchant 平台手动关闭/撤销用户入驻权限，撤销后仅可查看历史活动数据
func RevokeMerchant(adminID int64, adminName string, userID int64) error {
	user, err := getUserByID(userID)
	if err != nil {
		return err
	}
	if user.MerchantFlag == 1 && user.Role != model.RoleMerchant {
		return errcode.ErrBusiness.WithMsg("该用户暂无入驻权限")
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"role": model.RoleNormal, "merchant_flag": 1}).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	_ = SendMessage(userID, model.MsgTypeMerchant, "入驻权限已收回",
		"您的主办方入驻权限已被平台收回，无法再发布新活动，历史活动数据仍可查看。如有疑问请联系平台客服。", 0)
	RecordLog(adminID, adminName, "用户管理", "撤销入驻权限", "用户ID："+itoa(userID), "")
	return nil
}

// notifyMerchantAuditResult 入驻审核结果通知
func notifyMerchantAuditResult(a *model.MerchantApply, adminName string) {
	if a == nil {
		return
	}
	user, err := getUserByID(a.UserID)
	if err != nil {
		return
	}
	title := "入驻申请已通过"
	content := "恭喜您通过平台审核，已成功开通主办方权限，现在可以发布和管理自己的活动了。"
	short := "审核通过"
	if a.Status == model.ApplyStatusRejected {
		title = "入驻申请被驳回"
		content = "您的入驻申请未通过审核，原因：" + a.AuditRemark + "。修改后可重新提交申请。"
		short = "审核驳回"
	}
	_ = SendMessage(a.UserID, model.MsgTypeMerchant, title, content, a.ID)
	PushSubscribe(user.Openid, config.Get().Wechat.TmplMerchantAudit, wechat.SubscribeData{
		"thing1":  {Value: cutThing("主办方入驻申请")},
		"phrase2": {Value: short},
		"thing3":  {Value: cutThing(strDefault(a.AuditRemark, adminName))},
	})
}
