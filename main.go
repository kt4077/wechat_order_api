package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"activity/config"
	"activity/internal/model"
	"activity/internal/router"
	"activity/internal/service"
	"activity/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	cfgPath := flag.String("c", "", "配置文件路径，默认 config/config.yaml")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Printf("配置加载失败：%v\n", err)
		os.Exit(1)
	}

	gin.SetMode(cfg.Server.Mode)
	if err := logger.Init("./runtime/logs", "info"); err != nil {
		fmt.Printf("日志初始化失败：%v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Infof("开始初始化数据库：%s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	if err := model.Init(cfg); err != nil {
		logger.Errorf("数据库初始化失败：%v", err)
		os.Exit(1)
	}
	service.EnsureSuperAdmin()
	service.EnsureDefaultConfig()

	engine := router.New()
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        engine,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Infof("服务启动成功，监听端口 :%d，运行模式 %s", cfg.Server.Port, cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("服务异常退出：%v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("服务关闭异常：%v", err)
	}
	logger.Info("服务已安全退出")

}
