package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

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

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

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

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	UserExpire  int    `mapstructure:"user_expire"`
	AdminExpire int    `mapstructure:"admin_expire"`
}

type AESConfig struct {
	Key string `mapstructure:"key"`
	IV  string `mapstructure:"iv"`
}

type OSSConfig struct {
	Driver  string   `mapstructure:"driver"` // local | qiniu
	MaxSize int64    `mapstructure:"max_size"`
	Local   LocalOSS `mapstructure:"local"`
	Qiniu   QiniuOSS `mapstructure:"qiniu"`
}

type LocalOSS struct {
	Root   string `mapstructure:"root"`
	Domain string `mapstructure:"domain"`
}

type QiniuOSS struct {
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Domain    string `mapstructure:"domain"`
	Zone      string `mapstructure:"zone"`
	Prefix    string `mapstructure:"prefix"`
}

type WechatConfig struct {
	AppID             string      `mapstructure:"app_id"`
	AppSecret         string      `mapstructure:"app_secret"`
	LogLevel          string      `mapstructure:"log_level"`
	Cache             WechatCache `mapstructure:"cache"`
	TmplSignupSubmit  string      `mapstructure:"tmpl_signup_submit"`
	TmplSignupAudit   string      `mapstructure:"tmpl_signup_audit"`
	TmplMerchantAudit string      `mapstructure:"tmpl_merchant_audit"`
}

type WechatCache struct {
	Driver string     `mapstructure:"driver"` // memory | redis
	Prefix string     `mapstructure:"prefix"`
	Redis  RedisCache `mapstructure:"redis"`
}

type RedisCache struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type RateLimitConfig struct {
	Enable bool `mapstructure:"enable"`
	QPS    int  `mapstructure:"qps"`
	Burst  int  `mapstructure:"burst"`
}

type SuperAdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Nickname string `mapstructure:"nickname"`
}

var (
	global *Config
	once   sync.Once
)

func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		if path == "" {
			path = "config/config.yaml"
		}
		v.SetConfigFile(path)
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

func Get() *Config {
	if global == nil {
		if _, err := Load(""); err != nil {
			return &Config{}
		}
	}
	return global
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}
