// Package validate 提供常用数据格式校验与分页参数处理工具。
package validate

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	phoneReg  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	idCardReg = regexp.MustCompile(`^\d{17}[\dXx]$`)
	// 身份证加权因子
	idCardWeight = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	idCardCode   = []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
)

// IsPhone 校验 11 位有效手机号
func IsPhone(phone string) bool {
	return phoneReg.MatchString(strings.TrimSpace(phone))
}

// IsIDCard 校验 18 位身份证（含校验位）
func IsIDCard(id string) bool {
	id = strings.TrimSpace(id)
	if !idCardReg.MatchString(id) {
		return false
	}
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(id[i]-'0') * idCardWeight[i]
	}
	mod := sum % 11
	return strings.EqualFold(idCardCode[mod], string(id[17]))
}

// NormalizePage 规范化分页参数，限制单页最大条数，避免大结果集拖垮服务
func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// Trim 去除字符串首尾空格
func Trim(s string) string { return strings.TrimSpace(s) }

// ToFloat64 安全转换浮点
func ToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return math.NaN()
}

// ParseTime 多格式时间解析，返回零值表示解析失败
func ParseTime(s string) (time.Time, bool) {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", "2006/01/02 15:04:05", time.RFC3339}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
