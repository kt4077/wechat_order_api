package wechat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"activity/config"
	"activity/pkg/logger"

	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/basicService/subscribeMessage/request"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/power"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

type Session struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

var (
	app     *miniProgram.MiniProgram
	appOnce sync.Once
	initErr error
)

func App() (*miniProgram.MiniProgram, error) {
	appOnce.Do(func() {
		cfg := config.Get().Wechat
		if cfg.AppID == "" || cfg.AppSecret == "" {
			initErr = errors.New("微信配置缺失，请在 config/config.yaml 中填写 wechat.app_id 与 wechat.app_secret")
			return
		}
		level := cfg.LogLevel
		if level == "" {
			level = "info"
		}
		_ = os.MkdirAll("./runtime/logs", 0o755)
		app, initErr = miniProgram.NewMiniProgram(&miniProgram.UserConfig{
			AppID:  cfg.AppID,
			Secret: cfg.AppSecret,
			Log: miniProgram.Log{
				Level: level,
				File:  "./runtime/logs/wechat.log",
				ENV:   config.Get().Server.Mode,
			},
			Cache:     buildCache(),
			HttpDebug: config.Get().Server.Mode == "debug",
		})
		if initErr != nil {
			logger.Errorf("PowerWeChat 初始化失败：%v", initErr)
		}
	})
	return app, initErr
}

func buildCache() kernel.CacheInterface {
	c := config.Get().Wechat.Cache
	if c.Driver == "redis" && c.Redis.Host != "" {
		return kernel.NewRedisClient(&kernel.UniversalOptions{
			Addrs:    []string{c.Redis.Host + ":" + strconv.Itoa(c.Redis.Port)},
			Password: c.Redis.Password,
			DB:       c.Redis.DB,
		})
	}
	prefix := c.Prefix
	if prefix == "" {
		prefix = "service_api"
	}
	dir := "./runtime/cache"
	_ = os.MkdirAll(dir, 0o755)
	return cache.NewMemCache(prefix, 0, dir)
}

func Code2Session(code string) (*Session, error) {
	client, err := App()
	if err != nil {
		return nil, err
	}
	result, err := client.Auth.Session(context.Background(), code)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败(%d)：%s", result.ErrCode, result.ErrMsg)
	}
	if result.OpenID == "" {
		return nil, errors.New("微信登录失败：openid 为空")
	}
	return &Session{
		OpenID:     result.OpenID,
		SessionKey: result.SessionKey,
		UnionID:    result.UnionID,
	}, nil
}

func GetPhoneNumber(code string) (string, error) {
	client, err := App()
	if err != nil {
		return "", err
	}
	result, err := client.PhoneNumber.GetUserPhoneNumber(context.Background(), code)
	if err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取手机号失败(%d)：%s", result.ErrCode, result.ErrMsg)
	}
	if result.PhoneInfo == nil || result.PhoneInfo.PhoneNumber == "" {
		return "", errors.New("获取手机号失败：返回内容为空")
	}
	return result.PhoneInfo.PhoneNumber, nil
}

type SubscribeData map[string]struct {
	Value string `json:"value"`
}

func SendSubscribeMessage(openID, templateID string, data SubscribeData) error {
	if openID == "" || templateID == "" {
		return nil
	}
	client, err := App()
	if err != nil {
		logger.Warnf("发送订阅消息失败：%v", err)
		return nil
	}
	content := &power.HashMap{}
	for key, item := range data {
		(*content)[key] = power.HashMap{"value": item.Value}
	}
	result, err := client.SubscribeMessage.Send(context.Background(), &request.RequestSubscribeMessageSend{
		ToUser:     openID,
		TemplateID: templateID,
		Data:       content,
	})
	if err != nil {
		logger.Warnf("发送订阅消息异常：%v", err)
		return nil
	}
	if result != nil && result.ErrCode != 0 {
		logger.Warnf("发送订阅消息失败(%d)：%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
