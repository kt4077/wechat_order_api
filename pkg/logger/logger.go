// Package logger 基于 zap 封装全局日志，支持控制台输出与按天切割文件输出。
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global *zap.SugaredLogger
	once   sync.Once
)

// Init 初始化全局日志。dir 为空时仅输出到控制台。
func Init(dir string, level string) error {
	var err error
	once.Do(func() {
		var lvl zapcore.Level
		if err = lvl.UnmarshalText([]byte(level)); err != nil {
			lvl = zapcore.InfoLevel
			err = nil
		}
		cores := []zapcore.Core{}
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lvl))

		if dir != "" {
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				err = fmt.Errorf("创建日志目录失败：%w", mkErr)
				return
			}
			fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			ws, _, wErr := zap.Open(filepath.Join(dir, "app.log"))
			if wErr != nil {
				err = fmt.Errorf("打开日志文件失败：%w", wErr)
				return
			}
			cores = append(cores, zapcore.NewCore(fileEncoder, ws, lvl))
		}
		global = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	})
	return err
}

// Get 获取全局日志实例
func Get() *zap.SugaredLogger {
	if global == nil {
		_ = Init("", "info")
	}
	return global
}

// Info 普通日志
func Info(args ...interface{}) { Get().Info(args...) }

// Infof 格式化普通日志
func Infof(format string, args ...interface{}) { Get().Infof(format, args...) }

// Warn 警告日志
func Warn(args ...interface{}) { Get().Warn(args...) }

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) { Get().Warnf(format, args...) }

// Error 错误日志
func Error(args ...interface{}) { Get().Error(args...) }

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) { Get().Errorf(format, args...) }

// Fatal 致命错误日志
func Fatalf(format string, args ...interface{}) { Get().Fatalf(format, args...) }

// Sync 刷新缓冲
func Sync() { _ = Get().Sync() }
