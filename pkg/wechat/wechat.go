// Package wechat 基于 PowerWeChat 封装微信小程序服务端能力：
// code2session 登录、access_token 托管、订阅消息推送、手机号快速验证。
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

// Session code2session 登录凭证校验结果
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

// App 获取 PowerWeChat 小程序应用实例（进程内单例，内部自动托管 access_token 缓存）
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

// buildCache 构建 access_token 缓存：memory 使用进程内缓存，redis 适用于多实例部署
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
	// 目录必须预先存在，否则 PowerLibs 会退化到用户主目录
	dir := "./runtime/cache"
	_ = os.MkdirAll(dir, 0o755)
	return cache.NewMemCache(prefix, 0, dir)
}

// Code2Session 使用临时登录凭证 code 换取 openid 与 session_key
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

// GetPhoneNumber 解析微信手机号快速验证组件返回的 code
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

// SubscribeData 订阅消息内容项，键为模板关键词占位符（如 thing1、phrase2）
type SubscribeData map[string]struct {
	Value string `json:"value"`
}

// SendSubscribeMessage 发送微信订阅消息。
// 未配置模板 ID、用户未授权或微信侧报错时只记录日志，不影响业务主流程。
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
