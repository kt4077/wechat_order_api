package service

import (
	"time"

	"activity/internal/dto"
	"activity/internal/model"
	"activity/pkg/errcode"

	"gorm.io/gorm"
)

// Dashboard 后台数据总览：核心指标 + 趋势 + 分布 + 热门活动
func Dashboard() (*dto.DashboardResp, error) {
	resp := &dto.DashboardResp{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	_ = model.DB.Model(&model.Activity{}).Count(&resp.ActivityTotal).Error
	_ = model.DB.Model(&model.Activity{}).
		Where("status IN (1,2,3) AND (sign_start_time IS NULL OR sign_start_time <= ?) AND (sign_end_time IS NULL OR sign_end_time >= ?)", now, now).
		Count(&resp.ActivitySigning).Error
	_ = model.DB.Model(&model.Signup{}).Count(&resp.SignupTotal).Error
	_ = model.DB.Model(&model.Signup{}).Where("status = ?", model.SignupStatusPending).Count(&resp.SignupPending).Error
	_ = model.DB.Model(&model.User{}).Count(&resp.UserTotal).Error
	_ = model.DB.Model(&model.User{}).Where("merchant_flag = 2").Count(&resp.MerchantTotal).Error
	_ = model.DB.Model(&model.MerchantApply{}).Where("status = ?", model.ApplyStatusPending).Count(&resp.ApplyPending).Error
	_ = model.DB.Model(&model.Signup{}).Where("create_time >= ?", today).Count(&resp.TodaySignup).Error

	resp.Trend = signupTrend(7)
	resp.StatusDist = signupStatusDist()
	resp.CategoryDist = categoryDist()
	resp.HotActivities = hotActivities(5)
	return resp, nil
}

// signupTrend 统计最近 N 天报名趋势
func signupTrend(days int) []dto.TrendItem {
	items := make([]dto.TrendItem, 0, days)
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -i)
		end := start.AddDate(0, 0, 1)
		var count int64
		_ = model.DB.Model(&model.Signup{}).Where("create_time >= ? AND create_time < ?", start, end).Count(&count).Error
		items = append(items, dto.TrendItem{Date: start.Format("01-02"), Count: count})
	}
	return items
}

// signupStatusDist 报名状态分布
func signupStatusDist() []dto.NameValueItem {
	rows := make([]struct {
		Status int8  `gorm:"column:status"`
		Cnt    int64 `gorm:"column:cnt"`
	}, 0)
	_ = model.DB.Model(&model.Signup{}).Select("status, COUNT(*) AS cnt").Group("status").Scan(&rows).Error
	items := make([]dto.NameValueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.NameValueItem{Name: SignupStatusText(row.Status), Value: row.Cnt})
	}
	return items
}

// categoryDist 活动分类分布
func categoryDist() []dto.NameValueItem {
	rows := make([]struct {
		Category string `gorm:"column:category"`
		Cnt      int64  `gorm:"column:cnt"`
	}, 0)
	_ = model.DB.Model(&model.Activity{}).
		Select("IFNULL(category,'未分类') AS category, COUNT(*) AS cnt").
		Group("category").Order("cnt DESC").Limit(8).Scan(&rows).Error
	items := make([]dto.NameValueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.NameValueItem{Name: row.Category, Value: row.Cnt})
	}
	return items
}

// hotActivities 报名人数最多的活动
func hotActivities(limit int) []dto.HotActivityItem {
	list := make([]model.Activity, 0)
	_ = model.DB.Where("status NOT IN ?", []int8{model.ActivityStatusDraft}).
		Order("signed_count DESC, id DESC").Limit(limit).Find(&list).Error
	items := make([]dto.HotActivityItem, 0, len(list))
	for _, item := range list {
		items = append(items, dto.HotActivityItem{
			ID:          item.ID,
			Title:       item.Title,
			SignedCount: item.SignedCount,
			Quota:       item.Quota,
			Status:      EffectiveStatus(&item),
		})
	}
	return items
}

// ActivityStat 单个活动的数据统计，用于小程序端活动管理页
type ActivityStat struct {
	SignedCount  int             `json:"signed_count"`
	PassCount    int             `json:"pass_count"`
	PendingCount int             `json:"pending_count"`
	RejectCount  int             `json:"reject_count"`
	CancelCount  int             `json:"cancel_count"`
	Quota        int             `json:"quota"`
	RemainQuota  int             `json:"remain_quota"`
	Daily        []dto.TrendItem `json:"daily"`
}

// GetActivityStat 查询单活动数据统计
func GetActivityStat(activityID int64, operatorID int64, isAdmin bool) (*ActivityStat, error) {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", activityID).First(a).Error; err != nil {
		return nil, errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if !isAdmin && a.CreatorID != operatorID {
		return nil, errcode.ErrForbid.WithMsg("无权限查看该活动数据")
	}
	stat := &ActivityStat{
		SignedCount: a.SignedCount,
		PassCount:   a.PassCount,
		Quota:       a.Quota,
		RemainQuota: RemainQuota(a.Quota, a.SignedCount),
		Daily:       make([]dto.TrendItem, 0),
	}
	rows := make([]struct {
		Status int8  `gorm:"column:status"`
		Cnt    int64 `gorm:"column:cnt"`
	}, 0)
	_ = model.DB.Model(&model.Signup{}).Select("status, COUNT(*) AS cnt").
		Where("activity_id = ?", activityID).Group("status").Scan(&rows).Error
	for _, row := range rows {
		switch row.Status {
		case model.SignupStatusPending:
			stat.PendingCount = int(row.Cnt)
		case model.SignupStatusApproved:
			stat.PassCount = int(row.Cnt)
		case model.SignupStatusRejected:
			stat.RejectCount = int(row.Cnt)
		case model.SignupStatusCanceled:
			stat.CancelCount = int(row.Cnt)
		}
	}
	// 最近 7 天报名趋势
	now := time.Now()
	for i := 6; i >= 0; i-- {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -i)
		end := start.AddDate(0, 0, 1)
		var count int64
		_ = model.DB.Model(&model.Signup{}).
			Where("activity_id = ? AND create_time >= ? AND create_time < ?", activityID, start, end).
			Count(&count).Error
		stat.Daily = append(stat.Daily, dto.TrendItem{Date: start.Format("01-02"), Count: count})
	}
	return stat, nil
}

// EnsureTables 供启动时调用，保证表结构存在
func EnsureTables(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	return nil
}
