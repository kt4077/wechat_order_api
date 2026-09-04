package service

import (
	"fmt"
	"strconv"
	"strings"

	"activity/internal/model"
	"activity/pkg/errcode"
	"activity/pkg/validate"

	"github.com/xuri/excelize/v2"
)

func ExportSignups(activityID int64, status int, creatorID int64, isAdmin bool) ([]byte, string, error) {
	a := &model.Activity{}
	if err := model.DB.Where("id = ?", activityID).First(a).Error; err != nil {
		return nil, "", errcode.ErrNotFound.WithMsg("活动不存在")
	}
	if !isAdmin && a.CreatorID != creatorID {
		return nil, "", errcode.ErrForbid.WithMsg("仅活动创建者可导出报名数据")
	}

	tx := model.DB.Model(&model.Signup{}).Where("activity_id = ?", activityID)
	if status >= 0 {
		tx = tx.Where("status = ?", status)
	}
	list := make([]model.Signup, 0)
	if err := tx.Order("id ASC").Find(&list).Error; err != nil {
		return nil, "", errcode.ErrSystem.WithMsg("查询报名数据失败")
	}

	fields := ParseFormConfig(a.FormConfig)
	file := excelize.NewFile()
	sheet := "报名数据"
	index, err := file.NewSheet(sheet)
	if err != nil {
		return nil, "", errcode.ErrSystem.WithMsg("创建表格失败")
	}
	file.SetActiveSheet(index)
	_ = file.DeleteSheet("Sheet1")

	headers := []string{"序号", "报名编号", "报名人", "报名状态", "预约时间", "审核备注", "提交时间"}
	for _, field := range fields {
		headers = append(headers, field.Label)
	}
	headStyle, _ := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#2979FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
		_ = file.SetCellStyle(sheet, cell, cell, headStyle)
	}

	bodyStyle, _ := file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	for rowIdx, item := range list {
		row := rowIdx + 2
		data := ParseFormData(item.FormData)
		values := []string{
			strconv.Itoa(rowIdx + 1),
			fmt.Sprintf("%06d", item.ID),
			item.Nickname,
			SignupStatusText(item.Status),
			model.FmtTime(item.AppointTime),
			item.AuditRemark,
			model.FmtTimeValue(item.CreateTime),
		}
		for _, field := range fields {
			values = append(values, asString(data[field.Key]))
		}
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			_ = file.SetCellValue(sheet, cell, value)
			_ = file.SetCellStyle(sheet, cell, cell, bodyStyle)
		}
	}
	for i, header := range headers {
		width := 14.0
		if runeWidth := float64(len([]rune(header)) * 2); runeWidth > width {
			width = runeWidth
		}
		if width > 40 {
			width = 40
		}
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = file.SetColWidth(sheet, colName, colName, width)
	}
	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, "", errcode.ErrSystem.WithMsg("生成 Excel 失败")
	}
	fileName := fmt.Sprintf("活动报名数据_%d_%s.xlsx", activityID, validate.Trim(strings.ReplaceAll(a.Title, " ", "")))
	return buf.Bytes(), fileName, nil
}
