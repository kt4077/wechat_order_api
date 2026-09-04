package service

import (
	"activity/internal/dto"
	"activity/internal/model"
	"activity/pkg/errcode"
	"activity/pkg/validate"
)

func SendMessage(userID int64, msgType int8, title, content string, relatedID int64) error {
	if userID <= 0 {
		return nil
	}
	msg := &model.Message{
		UserID:    userID,
		Type:      msgType,
		Title:     title,
		Content:   content,
		RelatedID: relatedID,
		IsRead:    1,
	}
	if err := model.DB.Create(msg).Error; err != nil {
		return errcode.ErrSystem.WithMsg("写入消息失败")
	}
	return nil
}

func ListMessages(userID int64, q *dto.MessageQuery) ([]dto.MessageItem, int64, error) {
	page, pageSize := validate.NormalizePage(q.Page, q.PageSize)
	tx := model.DB.Model(&model.Message{}).Where("user_id = ?", userID)
	if q.Type > 0 {
		tx = tx.Where("type = ?", q.Type)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	list := make([]model.Message, 0)
	if err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errcode.ErrSystem
	}
	items := make([]dto.MessageItem, 0, len(list))
	for _, item := range list {
		items = append(items, dto.MessageItem{
			ID:         item.ID,
			Type:       item.Type,
			TypeText:   MessageTypeText(item.Type),
			Title:      item.Title,
			Content:    item.Content,
			RelatedID:  item.RelatedID,
			IsRead:     item.IsRead,
			CreateTime: model.FmtTimeValue(item.CreateTime),
		})
	}
	return items, total, nil
}

func ReadMessage(userID, id int64) error {
	if err := model.DB.Model(&model.Message{}).Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", 2).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	return nil
}

func ReadAllMessages(userID int64) error {
	if err := model.DB.Model(&model.Message{}).Where("user_id = ? AND is_read = 1", userID).
		Update("is_read", 2).Error; err != nil {
		return errcode.ErrSystem.WithMsg("操作失败")
	}
	return nil
}

func UnreadCount(userID int64) int64 {
	var count int64
	_ = model.DB.Model(&model.Message{}).Where("user_id = ? AND is_read = 1", userID).Count(&count).Error
	return count
}

func MessageTypeText(t int8) string {
	switch t {
	case model.MsgTypeSignup:
		return "报名通知"
	case model.MsgTypeAudit:
		return "审核通知"
	case model.MsgTypeMerchant:
		return "入驻通知"
	case model.MsgTypeActivity:
		return "活动通知"
	default:
		return "系统通知"
	}
}
