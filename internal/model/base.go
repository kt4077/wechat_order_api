// Package model 定义数据库实体与连接管理。所有业务表均包含 id / create_time / update_time / delete_time 通用字段。
package model

import (
	"time"

	"activity/config"
	"activity/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// BaseModel 数据库表通用字段，删除统一采用软删除
type BaseModel struct {
	ID         int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreateTime time.Time      `gorm:"column:create_time;autoCreateTime;index" json:"create_time"`
	UpdateTime time.Time      `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
	DeleteTime gorm.DeletedAt `gorm:"column:delete_time;index" json:"-"`
}

// DB 全局数据库句柄
var DB *gorm.DB

// Init 初始化数据库连接并执行自动迁移
func Init(cfg *config.Config) error {
	var level glogger.LogLevel
	switch cfg.Database.LogLevel {
	case "silent":
		level = glogger.Silent
	case "error":
		level = glogger.Error
	case "info":
		level = glogger.Info
	default:
		level = glogger.Warn
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       cfg.Database.DSN(),
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: glogger.Default.LogMode(level),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db

	if cfg.Database.AutoMigrate {
		if err := AutoMigrate(); err != nil {
			return err
		}
	}
	return nil
}

// AutoMigrate 自动建表并创建必要的索引
func AutoMigrate() error {
	err := DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务表'").AutoMigrate(
		&User{},
		&MerchantApply{},
		&Activity{},
		&Signup{},
		&Message{},
		&Admin{},
		&SysConfig{},
		&OperationLog{},
	)
	if err != nil {
		return err
	}
	return ensureExtraIndex()
}

// ensureExtraIndex 补充 AutoMigrate 未覆盖的复合索引
func ensureExtraIndex() error {
	type index struct {
		table string
		name  string
		field string
	}
	list := []index{
		{"activity", "idx_activity_status", "status"},
		{"activity", "idx_activity_creator", "creator_id"},
		{"activity", "idx_activity_official", "is_official"},
		{"signup", "idx_signup_activity", "activity_id"},
		{"signup", "idx_signup_user", "user_id"},
		{"signup", "idx_signup_status", "status"},
		{"merchant_apply", "idx_apply_user", "user_id"},
		{"merchant_apply", "idx_apply_status", "status"},
		{"message", "idx_message_user", "user_id"},
		{"operation_log", "idx_log_admin", "admin_id"},
	}
	for _, item := range list {
		if DB.Migrator().HasIndex(item.table, item.name) {
			continue
		}
		if err := DB.Exec("CREATE INDEX " + item.name + " ON `" + item.table + "` (`" + item.field + "`)").Error; err != nil {
			pkgLogWarn(err)
		}
	}
	return nil
}

func pkgLogWarn(err error) {
	logger.Warnf("创建索引跳过：%v", err)
}

// TimeFormat 统一时间输出格式
const TimeFormat = "2006-01-02 15:04:05"

// FmtTime 将时间指针格式化为字符串，nil 返回空串
func FmtTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}

// FmtTimeValue 将时间值格式化为字符串
func FmtTimeValue(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}
