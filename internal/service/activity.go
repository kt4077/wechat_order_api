package service

import (
	"errors"
	"fmt"
	"time"

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

func calcPersistStatus(a *model.Activity) int8 {
	now := time.Now()
	if a.SignStartTime != nil && now.Before(*a.SignStartTime) {
		return model.ActivityStatusNotStart
	}
	if a.SignEndTime != nil && now.After(*a.SignEndTime) {
		return model.ActivityStatusEnded
	}
	if a.EndTime != nil && now.After(*a.EndTime) {
		return model.ActivityStatusEnded
	}
	return model.ActivityStatusSigning
}

func EffectiveStatus(a *model.Activity) int8 {
	switch a.Status {
	case model.ActivityStatusDraft, model.ActivityStatusOffline,
		model.ActivityStatusPending, model.ActivityStatusRejected:
		return a.Status
	}
	return calcPersistStatus(a)
}

func StatusText(status int8, quota, signedCount int) string {
	switch status {
	case model.ActivityStatusDraft:
		return "草稿"
	case model.ActivityStatusNotStart:
		return "未开始"
	case model.ActivityStatusSigning:
		if quota > 0 && signedCount >= quota {
			return "名额已满"
		}
		return "报名中"
	case model.ActivityStatusEnded:
		return "已结束"
	case model.ActivityStatusOffline:
		return "已下架"
	case model.ActivityStatusPending:
		return "待审核"
	case model.ActivityStatusRejected:
		return "审核驳回"
	default:
		return "未知"
	}
}

func RemainQuota(quota, signedCount int) int {
	if quota <= 0 {
		return -1
	}
	remain := quota - signedCount
	if remain < 0 {
		remain = 0
	}
	return remain
}

func ListActivities(q *dto.ActivityQuery) ([]dto.ActivityListItem, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.Activity{})
	if q.Keyword != "" {
		tx = tx.Where("title LIKE ?", "%"+q.Keyword+"%")
	}
	if q.Category != "" {
		tx = tx.Where("category = ?", q.Category)
	}
	if q.Type > 0 {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.Official == 1 {
		tx = tx.Where("is_official = ?", model.ActivityOfficialYes)
	} else if q.Official == 2 {
		tx = tx.Where("is_official = ?", model.ActivityOfficialNo)
	}
	if q.Scope == "mine" {
		tx = tx.Where("creator_id = ?", q.UserID)
	}

	now := time.Now()
	switch q.Status {
	case "draft":
		tx = tx.Where("status = ?", model.ActivityStatusDraft)
	case "offline":
		tx = tx.Where("status = ?", model.ActivityStatusOffline)
	case "notstart":
		tx = tx.Where("status IN (?,?,?) AND sign_start_time > ?",
			model.ActivityStatusNotStart, model.ActivityStatusSigning, model.ActivityStatusEnded, now)
	case "signing":
		tx = tx.Where("status IN (?,?,?) AND (sign_start_time IS NULL OR sign_start_time <= ?) AND (sign_end_time IS NULL OR sign_end_time >= ?)",
			model.ActivityStatusNotStart, model.ActivityStatusSigning, model.ActivityStatusEnded, now, now)
	case "ended":
		tx = tx.Where("status = ? OR (status IN (?,?) AND (sign_end_time < ? OR end_time < ?))",
			model.ActivityStatusEnded, model.ActivityStatusNotStart, model.ActivityStatusSigning, now, now)
	case "pending":
		tx = tx.Where("status = ?", model.ActivityStatusPending)
	case "rejected":
		tx = tx.Where("status = ?", model.ActivityStatusRejected)
	default:
		if q.Scope != "mine" && q.Scope != "all" {
			tx = tx.Where("status NOT IN ?", []int8{
				model.ActivityStatusDraft, model.ActivityStatusOffline,
				model.ActivityStatusPending, model.ActivityStatusRejected,
			})
		}
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.Activity, 0)
	if err := tx.Order("is_official DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.ActivityListItem, 0, len(list))
	ids := make([]int64, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.ID)
		items = append(items, *BuildActivityListItem(&item))
	}
	pendingMap, _ := countPendingSignups(ids)
	for i := range items {
		items[i].PendingCount = int(pendingMap[items[i].ID])
	}
	return items, total, nil
}

func countPendingSignups(activityIDs []int64) (map[int64]int64, error) {
	result := map[int64]int64{}
	if len(activityIDs) == 0 {
		return result, nil
	}
	rows := make([]struct {
		ActivityID int64 `gorm:"column:activity_id"`
		Cnt        int64 `gorm:"column:cnt"`
	}, 0)
	if err := model.DB.Model(&model.Signup{}).
		Select("activity_id, COUNT(*) AS cnt").
		Where("activity_id IN ? AND status = ?", activityIDs, model.SignupStatusPending).
		Group("activity_id").Scan(&rows).Error; err != nil {
		return result, err
	}
	for _, row := range rows {
		result[row.ActivityID] = row.Cnt
	}
	return result, nil
}

func BuildActivityListItem(a *model.Activity) *dto.ActivityListItem {
	status := EffectiveStatus(a)
	return &dto.ActivityListItem{
		ID:            a.ID,
		Title:         a.Title,
		Cover:         a.Cover,
		Category:      a.Category,
		Type:          a.Type,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		Address:       a.Address,
		Latitude:      a.Latitude,
		Longitude:     a.Longitude,
		PoiName:       a.PoiName,
		SignStartTime: model.FmtTime(a.SignStartTime),
		SignEndTime:   model.FmtTime(a.SignEndTime),
		StartTime:     model.FmtTime(a.StartTime),
		EndTime:       model.FmtTime(a.EndTime),
		Quota:         a.Quota,
		SignedCount:   a.SignedCount,
		PassCount:     a.PassCount,
		RemainQuota:   RemainQuota(a.Quota, a.SignedCount),
		Status:        status,
		StatusText:    StatusText(status, a.Quota, a.SignedCount),
		IsOfficial:    a.IsOfficial,
		CreatorID:     a.CreatorID,
		CreatorName:   a.CreatorName,
		Organizer:     a.Organizer,
		ViewCount:     a.ViewCount,
		CreateTime:    model.FmtTimeValue(a.CreateTime),
	}
}

func ActivityDetail(id, userID int64) (*dto.ActivityDetail, error) {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", id).First(a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound.WithMsg("活动不存在或已下架")
		}
		return nil, errcode.ErrSystem
	}
	status := EffectiveStatus(a)
	detail := &dto.ActivityDetail{
		ActivityListItem: *BuildActivityListItem(a),
		Description:      a.Description,
		Notice:           a.Notice,
		FormConfig:       ParseFormConfig(a.FormConfig),
		ContactName:      a.ContactName,
		AutoAudit:        a.AutoAudit,
		CanManage:        a.CreatorID == userID && userID > 0,
	}

	detail.CanSignup, detail.SignupTip = CanSignup(a, status)
	if userID > 0 {
		s := &model.Signup{}
		if err := model.DB.Where("activity_id = ? AND user_id = ? AND status <> ?", id, userID, model.SignupStatusCanceled).
			Order("id DESC").First(s).Error; err == nil {
			detail.MySignupID = s.ID
			detail.MyStatus = s.Status
			detail.CanSignup = false
			switch s.Status {
			case model.SignupStatusPending:
				detail.SignupTip = "您已报名，等待审核"
			case model.SignupStatusApproved:
				detail.SignupTip = "报名已通过"
			case model.SignupStatusRejected:
				detail.SignupTip = "报名被驳回，可重新报名"
				detail.CanSignup = status == model.ActivityStatusSigning && RemainQuota(a.Quota, a.SignedCount) != 0
			}
		}
	}
	if err := model.DB.Model(&model.Activity{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		logger.Warnf("更新活动浏览量失败：%v", err)
	}
	return detail, nil
}

func CanSignup(a *model.Activity, status int8) (bool, string) {
	switch status {
	case model.ActivityStatusDraft:
		return false, "活动草稿中，暂未发布"
	case model.ActivityStatusNotStart:
		return false, "报名尚未开始"
	case model.ActivityStatusEnded:
		return false, "报名已结束"
	case model.ActivityStatusOffline:
		return false, "活动已下架"
	case model.ActivityStatusPending:
		return false, "活动审核中，暂不可报名"
	case model.ActivityStatusRejected:
		return false, "活动审核未通过"
	}
	if RemainQuota(a.Quota, a.SignedCount) == 0 {
		return false, "报名名额已满"
	}
	return true, ""
}

func SaveActivity(userID int64, isOfficial bool, req *dto.ActivitySaveReq) (int64, error) {
	req.Title = validate.Trim(req.Title)
	if req.Title == "" {
		return 0, errcode.ErrParams.WithMsg("请输入活动名称")
	}
	if len([]rune(req.Title)) > 60 {
		return 0, errcode.ErrParams.WithMsg("活动名称不能超过 60 个字")
	}
	if req.Type != model.ActivityTypeOnline && req.Type != model.ActivityTypeOffline {
		return 0, errcode.ErrParams.WithMsg("请选择活动形式")
	}
	signStart, ok1 := validate.ParseTime(req.SignStartTime)
	signEnd, ok2 := validate.ParseTime(req.SignEndTime)
	if !ok1 || !ok2 {
		return 0, errcode.ErrParams.WithMsg("请设置完整的报名起止时间")
	}
	if !signEnd.After(signStart) {
		return 0, errcode.ErrParams.WithMsg("报名结束时间必须晚于开始时间")
	}
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		t, ok := validate.ParseTime(req.StartTime)
		if !ok {
			return 0, errcode.ErrParams.WithMsg("活动开始时间格式不正确")
		}
		startTime = &t
	}
	if req.EndTime != "" {
		t, ok := validate.ParseTime(req.EndTime)
		if !ok {
			return 0, errcode.ErrParams.WithMsg("活动结束时间格式不正确")
		}
		endTime = &t
	}
	if startTime != nil && endTime != nil && !endTime.After(*startTime) {
		return 0, errcode.ErrParams.WithMsg("活动结束时间必须晚于开始时间")
	}
	if req.Type == model.ActivityTypeOffline && validate.Trim(req.Address) == "" {
		return 0, errcode.ErrParams.WithMsg("线下活动请填写活动地址")
	}
	if req.Quota < 0 {
		return 0, errcode.ErrParams.WithMsg("报名名额不能为负数")
	}
	if req.AutoAudit != 1 && req.AutoAudit != 2 {
		req.AutoAudit = 1
	}
	if req.ContactPhone != "" && !validate.IsPhone(req.ContactPhone) {
		return 0, errcode.ErrParams.WithMsg("主办方联系电话格式不正确")
	}

	fields, err := NormalizeFormConfig(req.FormConfig)
	if err != nil {
		return 0, err
	}
	formJSON, _ := json.Marshal(fields)

	contactPhone, err := crypto.Encrypt(validate.Trim(req.ContactPhone))
	if err != nil {
		return 0, errcode.ErrSystem.WithMsg("联系方式加密失败")
	}

	creator, err := getUserByID(userID)
	if err != nil {
		return 0, err
	}

	entity := &model.Activity{
		Title:         req.Title,
		Cover:         req.Cover,
		Category:      validate.Trim(req.Category),
		Type:          req.Type,
		Province:      validate.Trim(req.Province),
		City:          validate.Trim(req.City),
		District:      validate.Trim(req.District),
		Address:       validate.Trim(req.Address),
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		PoiName:       validate.Trim(req.PoiName),
		Description:   req.Description,
		Notice:        req.Notice,
		SignStartTime: &signStart,
		SignEndTime:   &signEnd,
		StartTime:     startTime,
		EndTime:       endTime,
		Quota:         req.Quota,
		Organizer:     validate.Trim(req.Organizer),
		ContactName:   validate.Trim(req.ContactName),
		ContactPhone:  contactPhone,
		AutoAudit:     req.AutoAudit,
		FormConfig:    string(formJSON),
		CreatorID:     userID,
		CreatorName:   creator.Nickname,
		IsOfficial:    boolToInt8(isOfficial),
	}
	entity.Status = calcPersistStatus(entity)

	if req.ID > 0 {
		old := &model.Activity{}
		if err := model.DB.Where("id = ?", req.ID).First(old).Error; err != nil {
			return 0, errcode.ErrNotFound.WithMsg("活动不存在")
		}
		if old.CreatorID != userID && !isOfficial {
			return 0, errcode.ErrForbid.WithMsg("仅活动创建者可修改该活动")
		}
		entity.ID = req.ID
		entity.SignedCount = old.SignedCount
		entity.PassCount = old.PassCount
		entity.ViewCount = old.ViewCount
		entity.IsOfficial = old.IsOfficial
		entity.WarnNotified = old.WarnNotified
		if req.Status == model.ActivityStatusDraft {
			entity.Status = model.ActivityStatusDraft
		} else if old.Status == model.ActivityStatusOffline {
			entity.Status = model.ActivityStatusOffline
		} else if old.Status == model.ActivityStatusDraft ||
			old.Status == model.ActivityStatusPending ||
			old.Status == model.ActivityStatusRejected {
			entity.Status = model.ActivityStatusPending
		}
		if err := model.DB.Model(&model.Activity{}).Where("id = ?", req.ID).
			Omit("id", "create_time", "creator_id", "creator_name", "is_official", "signed_count", "pass_count", "view_count").
			Updates(entity).Error; err != nil {
			logger.Errorf("更新活动失败：%v", err)
			return 0, errcode.ErrSystem.WithMsg("保存活动失败")
		}
		return req.ID, nil
	}

	entity.SignedCount = 0
	entity.PassCount = 0
	entity.Status = model.ActivityStatusPending
	if req.Status == model.ActivityStatusDraft {
		entity.Status = model.ActivityStatusDraft
	}
	if err := model.DB.Create(entity).Error; err != nil {
		logger.Errorf("创建活动失败：%v", err)
		return 0, errcode.ErrSystem.WithMsg("创建活动失败")
	}
	return entity.ID, nil
}

func NormalizeFormConfig(fields []model.FormField) ([]model.FormField, error) {
	if len(fields) == 0 {
		return model.DefaultFormConfig(), nil
	}
	if len(fields) > 30 {
		return nil, errcode.ErrParams.WithMsg("自定义字段不能超过 30 个")
	}
	exists := map[string]bool{}
	result := make([]model.FormField, 0, len(fields))
	allowed := map[string]bool{
		"text": true, "textarea": true, "phone": true, "idcard": true, "number": true,
		"radio": true, "checkbox": true, "select": true, "image": true, "date": true,
	}
	for i, field := range fields {
		field.Key = validate.Trim(field.Key)
		field.Label = validate.Trim(field.Label)
		if field.Key == "" {
			field.Key = fmt.Sprintf("field_%d", i+1)
		}
		if field.Label == "" {
			return nil, errcode.ErrParams.WithMsg("自定义字段名称不能为空")
		}
		if !allowed[field.Type] {
			return nil, errcode.ErrParams.WithMsg("字段「" + field.Label + "」类型不支持")
		}
		if exists[field.Key] {
			return nil, errcode.ErrParams.WithMsg("字段标识「" + field.Key + "」重复")
		}
		exists[field.Key] = true
		switch field.Type {
		case "radio", "checkbox", "select":
			if len(field.Options) == 0 {
				return nil, errcode.ErrParams.WithMsg("字段「" + field.Label + "」请配置选项")
			}
			if len(field.Options) > 50 {
				return nil, errcode.ErrParams.WithMsg("字段「" + field.Label + "」选项不能超过 50 个")
			}
		}
		if field.Sort == 0 {
			field.Sort = i + 1
		}
		result = append(result, field)
	}
	return result, nil
}

func ChangeActivityStatus(userID int64, isAdmin bool, req *dto.ActivityStatusReq) error {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", req.ID).First(a).Error; err != nil {
		return errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if !isAdmin && a.CreatorID != userID {
		return errcode.ErrForbid.WithMsg("无权限操作该活动")
	}
	target := req.Status
	if target == model.ActivityStatusOffline {
		if err := model.DB.Model(&model.Activity{}).Where("id = ?", req.ID).
			Update("status", model.ActivityStatusOffline).Error; err != nil {
			return errcode.ErrSystem.WithMsg("操作失败")
		}
		notifyActivityOffline(a)
		return nil
	}
	a.Status = model.ActivityStatusSigning
	target = calcPersistStatus(a)
	if err := model.DB.Model(&model.Activity{}).Where("id = ?", req.ID).
		Updates(map[string]interface{}{"status": target, "warn_notified": 1}).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	return nil
}

func AuditActivity(req *dto.ActivityAuditReq) error {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", req.ID).First(a).Error; err != nil {
		return errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if a.Status != model.ActivityStatusPending {
		return errcode.ErrBusiness.WithMsg("仅待审核的活动可审核")
	}
	remark := validate.Trim(req.Remark)
	if req.Status == dto.AuditActionReject {
		if remark == "" {
			return errcode.ErrParams.WithMsg("驳回时请填写驳回原因")
		}
		if err := model.DB.Model(&model.Activity{}).Where("id = ?", req.ID).
			Update("status", model.ActivityStatusRejected).Error; err != nil {
			return errcode.ErrSystem.WithMsg("操作失败")
		}
		_ = SendMessage(a.CreatorID, model.MsgTypeActivity, "活动审核未通过",
			"您发布的活动《"+a.Title+"》未通过平台审核，原因："+remark+"，请修改后重新提交。", a.ID)
		return nil
	}
	a.Status = model.ActivityStatusSigning
	target := calcPersistStatus(a)
	if err := model.DB.Model(&model.Activity{}).Where("id = ?", req.ID).
		Updates(map[string]interface{}{"status": target, "warn_notified": 1}).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	_ = SendMessage(a.CreatorID, model.MsgTypeActivity, "活动审核通过",
		"您发布的活动《"+a.Title+"》已通过平台审核，现已对外展示，用户可在报名时间内参与报名。", a.ID)
	return nil
}

func DeleteActivity(userID int64, isAdmin bool, id int64) error {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", id).First(a).Error; err != nil {
		return errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if !isAdmin && a.CreatorID != userID {
		return errcode.ErrForbid.WithMsg("无权限删除该活动")
	}
	if err := model.DB.Delete(a).Error; err != nil {
		return errcode.ErrSystem.WithMsg("删除失败")
	}
	return nil
}

func CopyActivity(userID int64, isAdmin bool, id int64) (int64, error) {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", id).First(a).Error; err != nil {
		return 0, errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if !isAdmin && a.CreatorID != userID {
		return 0, errcode.ErrForbid.WithMsg("无权限复刻该活动")
	}
	creator, err := getUserByID(userID)
	if err != nil {
		return 0, err
	}
	title := a.Title + "（副本）"
	if len([]rune(title)) > 60 {
		title = string([]rune(title)[:60])
	}
	copied := &model.Activity{
		Title:         title,
		Cover:         a.Cover,
		Category:      a.Category,
		Type:          a.Type,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		Address:       a.Address,
		Latitude:      a.Latitude,
		Longitude:     a.Longitude,
		PoiName:       a.PoiName,
		Description:   a.Description,
		Notice:        a.Notice,
		SignStartTime: a.SignStartTime,
		SignEndTime:   a.SignEndTime,
		StartTime:     a.StartTime,
		EndTime:       a.EndTime,
		Quota:         a.Quota,
		Organizer:     a.Organizer,
		ContactName:   a.ContactName,
		ContactPhone:  a.ContactPhone,
		AutoAudit:     a.AutoAudit,
		FormConfig:    a.FormConfig,
		CreatorID:     userID,
		CreatorName:   creator.Nickname,
		IsOfficial:    a.IsOfficial,
		Status:        model.ActivityStatusDraft,
	}
	if err := model.DB.Create(copied).Error; err != nil {
		return 0, errcode.ErrSystem.WithMsg("复刻失败")
	}
	return copied.ID, nil
}

func notifyActivityOffline(a *model.Activity) {
	ids := make([]int64, 0)
	_ = model.DB.Model(&model.Signup{}).
		Where("activity_id = ? AND status IN (?,?)", a.ID, model.SignupStatusPending, model.SignupStatusApproved).
		Pluck("user_id", &ids).Error
	for _, uid := range ids {
		_ = SendMessage(uid, model.MsgTypeActivity, "活动已下架",
			"您报名的活动《"+a.Title+"》已被主办方下架，如有疑问请联系主办方。", a.ID)
	}
}

func QuotaWarnCheck(a *model.Activity) {
	if a.Quota <= 0 || a.WarnNotified == 2 || a.SignedCount < a.Quota {
		return
	}
	remain := a.Quota - a.SignedCount
	if remain > a.Quota/10 {
		return
	}
	_ = SendMessage(a.CreatorID, model.MsgTypeActivity, "名额预警",
		fmt.Sprintf("活动《%s》报名名额即将满员，剩余 %d 个名额，请及时查看。", a.Title, remain), a.ID)
	_ = model.DB.Model(&model.Activity{}).Where("id = ?", a.ID).Update("warn_notified", 2).Error
}

func PushSubscribe(openID, templateID string, data wechat.SubscribeData) {
	if openID == "" || templateID == "" {
		return
	}
	if err := wechat.SendSubscribeMessage(openID, templateID, data); err != nil {
		logger.Warnf("推送订阅消息失败：%v", err)
	}
}

func boolToInt8(b bool) int8 {
	if b {
		return 2
	}
	return 1
}
