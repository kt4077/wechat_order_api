// Package config 负责加载并对外暴露项目运行配置，支持多环境配置文件与环境变量覆盖。
package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Config 全局配置对象
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	AES        AESConfig        `mapstructure:"aes"`
	OSS        OSSConfig        `mapstructure:"oss"`
	Wechat     WechatConfig     `mapstructure:"wechat"`
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"ratelimit"`
	SuperAdmin SuperAdminConfig `mapstructure:"super_admin"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig MySQL 配置
type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	Charset     string `mapstructure:"charset"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxOpen     int    `mapstructure:"max_open"`
	LogLevel    string `mapstructure:"log_level"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

// JWTConfig 令牌配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	UserExpire  int    `mapstructure:"user_expire"`
	AdminExpire int    `mapstructure:"admin_expire"`
}

// AESConfig 隐私数据加密配置
type AESConfig struct {
	Key string `mapstructure:"key"`
	IV  string `mapstructure:"iv"`
}

// OSSConfig 对象存储配置，支持本地磁盘与七牛云
type OSSConfig struct {
	Driver  string   `mapstructure:"driver"` // local | qiniu
	MaxSize int64    `mapstructure:"max_size"`
	Local   LocalOSS `mapstructure:"local"`
	Qiniu   QiniuOSS `mapstructure:"qiniu"`
}

// LocalOSS 本地磁盘存储配置
type LocalOSS struct {
	Root   string `mapstructure:"root"`
	Domain string `mapstructure:"domain"`
}

// QiniuOSS 七牛云对象存储配置
type QiniuOSS struct {
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Domain    string `mapstructure:"domain"`
	Zone      string `mapstructure:"zone"`
	Prefix    string `mapstructure:"prefix"`
}

// WechatConfig 微信小程序配置（PowerWeChat）
type WechatConfig struct {
	AppID             string      `mapstructure:"app_id"`
	AppSecret         string      `mapstructure:"app_secret"`
	LogLevel          string      `mapstructure:"log_level"`
	Cache             WechatCache `mapstructure:"cache"`
	TmplSignupSubmit  string      `mapstructure:"tmpl_signup_submit"`
	TmplSignupAudit   string      `mapstructure:"tmpl_signup_audit"`
	TmplMerchantAudit string      `mapstructure:"tmpl_merchant_audit"`
}

// WechatCache 微信 access_token 缓存配置
type WechatCache struct {
	Driver string     `mapstructure:"driver"` // memory | redis
	Prefix string     `mapstructure:"prefix"`
	Redis  RedisCache `mapstructure:"redis"`
}

// RedisCache Redis 缓存连接配置
type RedisCache struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enable bool `mapstructure:"enable"`
	QPS    int  `mapstructure:"qps"`
	Burst  int  `mapstructure:"burst"`
}

// SuperAdminConfig 初始化超级管理员配置
type SuperAdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Nickname string `mapstructure:"nickname"`
}

var (
	global *Config
	once   sync.Once
)

// Load 加载配置。优先读取环境变量 ACTIVITY_ENV 指定的配置文件，默认 config.yaml。
func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		if path == "" {
			path = "config/config.yaml"
		}
		v.SetConfigFile(path)
		// 环境独立配置：config.{env}.yaml
		if env := os.Getenv("ACTIVITY_ENV"); env != "" {
			v.SetConfigName("config." + env)
			v.AddConfigPath("config")
			v.AddConfigPath(".")
		}
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		v.SetDefault("server.port", 8080)
		v.SetDefault("server.mode", "debug")
		v.SetDefault("database.charset", "utf8mb4")
		v.SetDefault("oss.driver", "qiniu")
		v.SetDefault("oss.max_size", 5)
		v.SetDefault("oss.local.root", "./uploads")
		v.SetDefault("oss.qiniu.prefix", "activity")
		v.SetDefault("wechat.log_level", "info")
		v.SetDefault("wechat.cache.driver", "memory")
		v.SetDefault("wechat.cache.prefix", "service_api")

		if err = v.ReadInConfig(); err != nil {
			err = fmt.Errorf("读取配置文件失败：%w", err)
			return
		}
		c := &Config{}
		if err = v.Unmarshal(c); err != nil {
			err = fmt.Errorf("解析配置文件失败：%w", err)
			return
		}
		global = c
	})
	return global, err
}

// Get 获取全局配置，未加载时返回零值配置，便于单元测试调用。
func Get() *Config {
	if global == nil {
		if _, err := Load(""); err != nil {
			return &Config{}
		}
	}
	return global
}

// DSN 生成 MySQL 连接串
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}
