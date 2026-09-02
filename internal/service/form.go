// Package service 承载全部业务逻辑，Controller 层仅负责参数接收与结果返回。
package service

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"activity/internal/model"
	"activity/pkg/errcode"
	"activity/pkg/validate"

	"encoding/json"
)

// asString 将任意表单值统一转换为字符串，便于统一校验与导出
func asString(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, asString(item))
		}
		return strings.Join(parts, ",")
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	}
}

// ValidateFormData 依据活动自定义表单配置校验报名数据
// 校验规则：必填、手机号格式、身份证格式、数字范围、选项合法性、文本长度
func ValidateFormData(fields []model.FormField, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	for _, field := range fields {
		val := asString(data[field.Key])
		if strings.TrimSpace(val) == "" {
			if field.Required {
				return errcode.ErrParams.WithMsg("请填写「" + field.Label + "」")
			}
			continue
		}
		switch field.Type {
		case "phone":
			if !validate.IsPhone(val) {
				return errcode.ErrParams.WithMsg("「" + field.Label + "」手机号格式不正确")
			}
		case "idcard":
			if !validate.IsIDCard(val) {
				return errcode.ErrParams.WithMsg("「" + field.Label + "」身份证号格式不正确")
			}
		case "number":
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return errcode.ErrParams.WithMsg("「" + field.Label + "」请输入数字")
			}
			if field.MinValue != nil && num < *field.MinValue {
				return errcode.ErrParams.WithMsg(fmt.Sprintf("「%s」不能小于 %s", field.Label, strconv.FormatFloat(*field.MinValue, 'f', -1, 64)))
			}
			if field.MaxValue != nil && num > *field.MaxValue {
				return errcode.ErrParams.WithMsg(fmt.Sprintf("「%s」不能大于 %s", field.Label, strconv.FormatFloat(*field.MaxValue, 'f', -1, 64)))
			}
		case "radio", "select":
			if !inOptions(field.Options, val) {
				return errcode.ErrParams.WithMsg("「" + field.Label + "」选项不合法")
			}
		case "checkbox":
			for _, item := range strings.Split(val, ",") {
				if !inOptions(field.Options, strings.TrimSpace(item)) {
					return errcode.ErrParams.WithMsg("「" + field.Label + "」选项不合法")
				}
			}
		}
		if field.MinLen > 0 && utf8.RuneCountInString(val) < field.MinLen {
			return errcode.ErrParams.WithMsg(fmt.Sprintf("「%s」至少输入 %d 个字", field.Label, field.MinLen))
		}
		if field.MaxLen > 0 && utf8.RuneCountInString(val) > field.MaxLen {
			return errcode.ErrParams.WithMsg(fmt.Sprintf("「%s」最多输入 %d 个字", field.Label, field.MaxLen))
		}
	}
	return nil
}

func inOptions(options []string, val string) bool {
	for _, opt := range options {
		if opt == val {
			return true
		}
	}
	return false
}

// ParseFormConfig 解析活动表单配置，解析失败时回退默认模板
func ParseFormConfig(raw string) []model.FormField {
	fields := make([]model.FormField, 0)
	if raw == "" {
		return model.DefaultFormConfig()
	}
	if err := json.Unmarshal([]byte(raw), &fields); err != nil || len(fields) == 0 {
		return model.DefaultFormConfig()
	}
	return fields
}

// ParseFormData 解析报名提交的表单数据
func ParseFormData(raw string) map[string]interface{} {
	data := map[string]interface{}{}
	if raw == "" {
		return data
	}
	_ = json.Unmarshal([]byte(raw), &data)
	return data
}
