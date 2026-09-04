package service

import (
	"strconv"
	"time"

	"gorm.io/gorm/clause"
)

func nowTime() time.Time { return time.Now() }

func nowPtr() *time.Time {
	t := nowTime()
	return &t
}

func clauseForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

func itoa(i int64) string { return strconv.FormatInt(i, 10) }
