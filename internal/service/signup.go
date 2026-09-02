package service

import (
	"activity/config"
	"activity/internal/dto"
	"activity/internal/model"
	"activity/pkg/crypto"
	"activity/pkg/errcode"
	"activity/pkg/validate"
	"activity/pkg/wechat"

	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SignupStatusText 报名状态文案
func SignupStatusText(status int8) string {
	switch status {
	case model.SignupStatusPending:
		return "待审核"
	case model.SignupStatusApproved:
		return "审核通过"
	case model.SignupStatusRejected:
		return "审核驳回"
	case model.SignupStatusCanceled:
		return "已撤销"
	default:
		return "未知"
	}
}

// SubmitSignup 提交报名。方法内使用事务加行锁，保证名额不超卖、单人单活动仅一次。
func SubmitSignup(userID int64, req *dto.SignupSubmitReq, clientIP string) (int64, error) {
	if req.ActivityID <= 0 {
		return 0, errcode.ErrParams.WithMsg("参数错误")
	}
	user, err := getUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user.Status == model.UserStatusDisable {
		return 0, errcode.ErrForbid.WithMsg("账号已被禁用，无法报名")
	}
	// 预约到场时间：报名者自行填写（选填）
	var appointTime *time.Time
	if validate.Trim(req.AppointTime) != "" {
		t, ok := validate.ParseTime(req.AppointTime)
		if !ok {
			return 0, errcode.ErrParams.WithMsg("预约时间格式不正确")
		}
		appointTime = &t
	}

	var signupID int64
	txErr := model.DB.Transaction(func(tx *gorm.DB) error {
		// 行锁防止并发超卖
		a := &model.Activity{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ActivityID).First(a).Error; err != nil {
			return errcode.ErrNotFound.WithMsg("活动不存在")
		}
		status := EffectiveStatus(a)
		if status != model.ActivityStatusSigning {
			switch status {
			case model.ActivityStatusNotStart:
				return errcode.ErrNotInTime
			case model.ActivityStatusEnded:
				return errcode.ErrEnded
			case model.ActivityStatusPending:
				return errcode.ErrBusiness.WithMsg("活动审核中，暂不可报名")
			case model.ActivityStatusRejected:
				return errcode.ErrBusiness.WithMsg("活动审核未通过")
			default:
				return errcode.ErrBusiness.WithMsg("该活动当前不可报名")
			}
		}
		if RemainQuota(a.Quota, a.SignedCount) == 0 {
			return errcode.ErrNoQuota
		}
		// 驳回记录允许重新报名：先归档为已撤销，保证同一时间仅存在一条有效记录
		now := nowTime()
		if err := tx.Model(&model.Signup{}).
			Where("activity_id = ? AND user_id = ? AND status = ?", a.ID, userID, model.SignupStatusRejected).
			Updates(map[string]interface{}{"status": model.SignupStatusCanceled, "cancel_time": now}).Error; err != nil {
			return errcode.ErrSystem.WithMsg("归档历史报名记录失败")
		}

		// 重复报名校验：待审核与已通过的记录占用名额，禁止重复提交
		var existCount int64
		if err := tx.Model(&model.Signup{}).
			Where("activity_id = ? AND user_id = ? AND status <> ?", a.ID, userID, model.SignupStatusCanceled).
			Count(&existCount).Error; err != nil {
			return errcode.ErrSystem
		}
		if existCount > 0 {
			return errcode.ErrSigned
		}

		fields := ParseFormConfig(a.FormConfig)
		if err := ValidateFormData(fields, req.FormData); err != nil {
			return err
		}

		formJSON, _ := json.Marshal(req.FormData)
		signup := &model.Signup{
			ActivityID:    a.ID,
			ActivityTitle: a.Title,
			ActivityCover: a.Cover,
			ActivityType:  a.Type,
			StartTime:     a.StartTime,
			EndTime:       a.EndTime,
			Address:       a.Address,
			UserID:        userID,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			FormData:      string(formJSON),
			AppointTime:   appointTime,
			Status:        model.SignupStatusPending,
			ClientIP:      clientIP,
		}
		updates := map[string]interface{}{"signed_count": gorm.Expr("signed_count + 1")}
		if a.AutoAudit == 2 {
			signup.Status = model.SignupStatusApproved
			signup.AuditRemark = "活动设置为免审核，自动通过"
			updates["pass_count"] = gorm.Expr("pass_count + 1")
		}
		if err := tx.Create(signup).Error; err != nil {
			return errcode.ErrSystem.WithMsg("提交报名失败")
		}
		if err := tx.Model(&model.Activity{}).Where("id = ?", a.ID).Updates(updates).Error; err != nil {
			return errcode.ErrSystem.WithMsg("更新报名数量失败")
		}
		signupID = signup.ID

		// 名额预警（在事务外处理，此处仅记录快照）
		a.SignedCount++
		go QuotaWarnCheck(a)
		return nil
	})
	if txErr != nil {
		return 0, txErr
	}

	// 报名提交成功通知
	_ = SendMessage(userID, model.MsgTypeSignup, "报名提交成功",
		"您已成功提交报名，请耐心等待主办方审核。", req.ActivityID)
	PushSubscribe(user.Openid, config.Get().Wechat.TmplSignupSubmit, wechat.SubscribeData{
		"thing1": {Value: cutThing(getActivityTitle(req.ActivityID))},
		"time2":  {Value: model.FmtTime(nowPtr())},
	})
	return signupID, nil
}

// ListSignups 报名记录列表，支持「我的报名」与「活动报名管理」两种场景
func ListSignups(q *dto.SignupQuery, isAdmin bool) ([]dto.SignupListItem, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.Signup{})
	if q.ActivityID > 0 {
		tx = tx.Where("activity_id = ?", q.ActivityID)
	}
	if q.Scope == "mine" {
		tx = tx.Where("user_id = ?", q.UserID)
	}
	if q.Status != nil && *q.Status >= 0 {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Keyword != "" {
		tx = tx.Where("(nickname LIKE ? OR activity_title LIKE ? OR form_data LIKE ?)",
			"%"+q.Keyword+"%", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.Signup, 0)
	if err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.SignupListItem, 0, len(list))
	for _, item := range list {
		items = append(items, *BuildSignupItem(&item, isAdmin))
	}
	return items, total, nil
}

// BuildSignupItem 组装报名列表项，脱敏开关控制手机号展示
func BuildSignupItem(s *model.Signup, showFull bool) *dto.SignupListItem {
	data := ParseFormData(s.FormData)
	if !showFull {
		// 非管理场景对隐私字段脱敏
		data = maskPrivacy(data)
	}
	return &dto.SignupListItem{
		ID:            s.ID,
		ActivityID:    s.ActivityID,
		ActivityTitle: s.ActivityTitle,
		ActivityCover: s.ActivityCover,
		ActivityType:  s.ActivityType,
		StartTime:     model.FmtTime(s.StartTime),
		EndTime:       model.FmtTime(s.EndTime),
		Address:       s.Address,
		UserID:        s.UserID,
		Nickname:      s.Nickname,
		Avatar:        s.Avatar,
		FormData:      data,
		Status:        s.Status,
		StatusText:    SignupStatusText(s.Status),
		AuditRemark:   s.AuditRemark,
		AuditorName:   s.AuditorName,
		AuditTime:     model.FmtTime(s.AuditTime),
		AppointTime:   model.FmtTime(s.AppointTime),
		CancelTime:    model.FmtTime(s.CancelTime),
		CreateTime:    model.FmtTimeValue(s.CreateTime),
	}
}

// maskPrivacy 对表单数据中的隐私字段做脱敏处理
func maskPrivacy(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for key, val := range data {
		str := asString(val)
		result[key] = val
		if len(str) == 11 && validate.IsPhone(str) {
			result[key] = crypto.MaskPhone(str)
			continue
		}
		if len(str) == 18 && validate.IsIDCard(str) {
			result[key] = crypto.MaskIDCard(str)
		}
	}
	return result
}

// SignupDetail 报名详情
func SignupDetail(id, userID int64, isAdmin bool) (*dto.SignupDetail, error) {
	s := &model.Signup{}
	if err := model.DB.Where("id = ?", id).First(s).Error; err != nil {
		return nil, errcode.ErrNotFound.WithMsg("报名记录不存在")
	}
	if !isAdmin && s.UserID != userID {
		return nil, errcode.ErrForbid.WithMsg("无权限查看该报名记录")
	}
	a := &model.Activity{}
	_ = model.DB.Where("id = ?", s.ActivityID).First(a).Error
	detail := &dto.SignupDetail{
		SignupListItem: *BuildSignupItem(s, isAdmin || s.UserID == userID),
		FormConfig:     ParseFormConfig(a.FormConfig),
	}
	if a.ID > 0 {
		item := BuildActivityListItem(a)
		detail.Activity = item
	}
	return detail, nil
}

// CancelSignup 取消报名：待审核与已通过均可取消，取消后释放名额。
// 已通过报名被取消时，同步扣减活动通过人数 pass_count。
func CancelSignup(userID, id int64) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		s := &model.Signup{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(s).Error; err != nil {
			return errcode.ErrNotFound.WithMsg("报名记录不存在")
		}
		if s.UserID != userID {
			return errcode.ErrForbid.WithMsg("无权限操作该报名记录")
		}
		if s.Status != model.SignupStatusPending && s.Status != model.SignupStatusApproved {
			return errcode.ErrBusiness.WithMsg("当前状态的报名不可取消")
		}
		now := nowTime()
		if err := tx.Model(&model.Signup{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":      model.SignupStatusCanceled,
				"cancel_time": now,
			}).Error; err != nil {
			return errcode.ErrSystem.WithMsg("取消报名失败")
		}
		actUpdates := map[string]interface{}{
			"signed_count": gorm.Expr("signed_count - 1"),
		}
		if s.Status == model.SignupStatusApproved {
			actUpdates["pass_count"] = gorm.Expr("pass_count - 1")
		}
		if err := tx.Model(&model.Activity{}).Where("id = ?", s.ActivityID).Updates(actUpdates).Error; err != nil {
			return errcode.ErrSystem.WithMsg("释放名额失败")
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = SendMessage(userID, model.MsgTypeSignup, "报名已取消",
		"您已成功取消报名，名额已释放，可重新报名参加其他活动。", id)
	return nil
}

// AuditSignup 审核报名：通过或驳回，并同步名额统计。
// 入参 req.Status 为「审核动作」（1通过 / 2驳回），内部映射为存储状态。
func AuditSignup(operatorID int64, operatorName string, isAdmin bool, req *dto.SignupAuditReq) error {
	if req.Status != dto.AuditActionPass && req.Status != dto.AuditActionReject {
		return errcode.ErrParams.WithMsg("审核状态不合法")
	}
	if req.Status == dto.AuditActionReject && validate.Trim(req.Remark) == "" {
		return errcode.ErrParams.WithMsg("驳回时请填写驳回原因")
	}
	// 动作 → 存储状态
	targetStatus := model.SignupStatusApproved
	if req.Status == dto.AuditActionReject {
		targetStatus = model.SignupStatusRejected
	}
	var target *model.Signup
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		s := &model.Signup{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).First(s).Error; err != nil {
			return errcode.ErrNotFound.WithMsg("报名记录不存在")
		}
		if !isAdmin {
			a := &model.Activity{}
			if err := tx.Where("id = ?", s.ActivityID).First(a).Error; err != nil {
				return errcode.ErrNotFound.WithMsg("活动不存在")
			}
			if a.CreatorID != operatorID {
				return errcode.ErrForbid.WithMsg("仅活动创建者可审核该报名")
			}
		}
		if s.Status != model.SignupStatusPending {
			return errcode.ErrBusiness.WithMsg("该报名已审核，请勿重复操作")
		}
		now := nowTime()
		updates := map[string]interface{}{
			"status":       targetStatus,
			"audit_remark": validate.Trim(req.Remark),
			"auditor_id":   operatorID,
			"auditor_name": operatorName,
			"audit_time":   now,
		}
		if err := tx.Model(&model.Signup{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			return errcode.ErrSystem.WithMsg("审核失败")
		}
		actUpdates := map[string]interface{}{}
		if targetStatus == model.SignupStatusApproved {
			actUpdates["pass_count"] = gorm.Expr("pass_count + 1")
		} else {
			actUpdates["signed_count"] = gorm.Expr("signed_count - 1")
		}
		if err := tx.Model(&model.Activity{}).Where("id = ?", s.ActivityID).Updates(actUpdates).Error; err != nil {
			return errcode.ErrSystem.WithMsg("更新统计失败")
		}
		s.Status = targetStatus
		s.AuditRemark = req.Remark
		target = s
		return nil
	})
	if err != nil {
		return err
	}
	notifySignupAuditResult(target, operatorName)
	return nil
}

// BatchAuditSignups 批量审核报名
func BatchAuditSignups(operatorID int64, operatorName string, isAdmin bool, req *dto.SignupBatchAuditReq) (int, error) {
	if len(req.IDs) == 0 {
		return 0, errcode.ErrParams.WithMsg("请选择要审核的记录")
	}
	success := 0
	var firstErr error
	for _, id := range req.IDs {
		if err := AuditSignup(operatorID, operatorName, isAdmin, &dto.SignupAuditReq{
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

// notifySignupAuditResult 报名审核结果通知（站内消息 + 微信订阅消息）
func notifySignupAuditResult(s *model.Signup, operatorName string) {
	if s == nil {
		return
	}
	user, err := getUserByID(s.UserID)
	if err != nil {
		return
	}
	title := "报名审核通过"
	content := "恭喜！您报名的活动《" + s.ActivityTitle + "》已审核通过，请按时参加。"
	short := "审核通过"
	if s.Status == model.SignupStatusRejected {
		title = "报名审核驳回"
		content = "很抱歉，您报名的活动《" + s.ActivityTitle + "》未通过审核，原因：" + s.AuditRemark
		short = "审核驳回"
	} else if s.AppointTime != nil {
		content = "恭喜！您报名的活动《" + s.ActivityTitle + "》已审核通过，预约时间：" + model.FmtTime(s.AppointTime)
	}
	_ = SendMessage(s.UserID, model.MsgTypeAudit, title, content, s.ID)
	PushSubscribe(user.Openid, config.Get().Wechat.TmplSignupAudit, wechat.SubscribeData{
		"thing1":  {Value: cutThing(s.ActivityTitle)},
		"phrase2": {Value: short},
		"thing3":  {Value: cutThing(strDefault(s.AuditRemark, operatorName))},
	})
}

func strDefault(s, def string) string {
	if validate.Trim(s) == "" {
		return def
	}
	return s
}

// cutThing 微信订阅消息 thing 类型最多 20 个字符
func cutThing(s string) string {
	runes := []rune(s)
	if len(runes) > 20 {
		return string(runes[:19]) + "…"
	}
	if s == "" {
		return "-"
	}
	return s
}

// getActivityTitle 查询活动标题，失败返回默认文案
func getActivityTitle(id int64) string {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", id).First(a).Error; err != nil {
		return "活动报名"
	}
	return a.Title
}
