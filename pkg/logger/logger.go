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

func Get() *zap.SugaredLogger {
	if global == nil {
		_ = Init("", "info")
	}
	return global
}

func Info(args ...interface{}) { Get().Info(args...) }

func Infof(format string, args ...interface{}) { Get().Infof(format, args...) }

func Warn(args ...interface{}) { Get().Warn(args...) }

func Warnf(format string, args ...interface{}) { Get().Warnf(format, args...) }

func Error(args ...interface{}) { Get().Error(args...) }

func Errorf(format string, args ...interface{}) { Get().Errorf(format, args...) }

func Fatalf(format string, args ...interface{}) { Get().Fatalf(format, args...) }

func Sync() { _ = Get().Sync() }
