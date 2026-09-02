package service

import (
	"strconv"
	"time"

	"gorm.io/gorm/clause"
)

// nowTime 统一时间入口，便于测试时替换为固定时间
func nowTime() time.Time { return time.Now() }

// nowPtr 返回当前时间指针
func nowPtr() *time.Time {
	t := nowTime()
	return &t
}

// clauseForUpdate 行级排他锁，用于报名、审核等并发写场景避免超卖与重复操作
func clauseForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

// itoa int64 转字符串
func itoa(i int64) string { return strconv.FormatInt(i, 10) }
