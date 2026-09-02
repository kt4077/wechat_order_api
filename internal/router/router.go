// Package router 统一管理路由分组，区分小程序用户端、PC 管理端与公共接口，实现权限隔离。
package router

import (
	"strings"

	"activity/config"
	"activity/internal/controller"
	"activity/internal/middleware"

	"github.com/gin-gonic/gin"
)

// New 初始化并注册全部路由
func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

	// 本地存储模式下通过静态目录对外提供上传文件访问；七牛云模式由 CDN 域名直接访问
	cfg := config.Get()
	if !strings.EqualFold(cfg.OSS.Driver, "qiniu") {
		root := cfg.OSS.Local.Root
		if root == "" {
			root = "./uploads"
		}
		r.Static("/uploads", root)
	}
	r.GET("/health", controller.Health)

	api := r.Group("/api")
	registerCommon(api)
	registerUser(api)
	registerAdmin(api)
	return r
}

// registerCommon 公共接口（无需登录）
func registerCommon(g *gin.RouterGroup) {
	common := g.Group("/common")
	{
		common.POST("/login", controller.Login)
		common.GET("/config", controller.GetConfig)
	}
}

// registerUser 小程序用户端接口
func registerUser(g *gin.RouterGroup) {
	// 活动浏览：支持游客访问，携带令牌时返回个人报名状态
	public := g.Group("")
	public.Use(middleware.OptionalUserAuth())
	{
		public.GET("/activities", controller.ListActivities)
		public.GET("/activities/:id", controller.ActivityDetail)
	}

	auth := g.Group("")
	auth.Use(middleware.UserAuth())
	{
		// 个人中心
		auth.GET("/user/profile", controller.GetProfile)
		auth.PUT("/user/profile", controller.UpdateProfile)
		auth.DELETE("/user/data", controller.DeleteMyData)

		// 消息通知
		auth.GET("/messages", controller.ListMessages)
		auth.GET("/messages/unread", controller.UnreadCount)
		auth.PUT("/messages/read", controller.ReadMessage)
		auth.PUT("/messages/read-all", controller.ReadAllMessages)

		// 图片上传
		auth.POST("/common/upload", controller.UploadImage)

		// 入驻申请
		auth.POST("/merchant/apply", controller.ApplyMerchant)
		auth.GET("/merchant/status", controller.MyMerchantStatus)

		// 活动管理（需入驻权限，控制器内二次校验）
		auth.POST("/activities", controller.CreateActivity)
		auth.PUT("/activities/:id", controller.UpdateActivity)
		auth.DELETE("/activities/:id", controller.DeleteActivity)
		auth.POST("/activities/:id/copy", controller.CopyActivity)
		auth.PUT("/activities/:id/status", controller.ChangeActivityStatus)
		auth.GET("/activities/:id/stat", controller.ActivityStat)
		auth.GET("/activities/:id/signups", controller.ListSignups)
		auth.GET("/activities/:id/signups/export", controller.ExportSignups)

		// 报名
		auth.POST("/signups", controller.SubmitSignup)
		auth.GET("/signups", controller.ListSignups)
		auth.GET("/signups/:id", controller.SignupDetail)
		auth.POST("/signups/:id/cancel", controller.CancelSignup)
		auth.PUT("/signups/:id/audit", controller.AuditSignup)
		auth.POST("/batch-audit-signups", controller.BatchAuditSignups)
	}
}

// registerAdmin PC 管理端接口
func registerAdmin(g *gin.RouterGroup) {
	admin := g.Group("/admin")
	{
		admin.POST("/login", controller.AdminLogin)

		auth := admin.Group("")
		auth.Use(middleware.AdminAuth())
		{
			auth.GET("/profile", controller.AdminProfile)
			auth.PUT("/password", controller.ChangeAdminPassword)
			auth.GET("/dashboard", controller.Dashboard)
			auth.POST("/upload", controller.AdminUpload)

			// 系统配置
			auth.GET("/config", controller.GetSysConfig)
			auth.PUT("/config", controller.SaveSysConfig)

			// 权限管理
			auth.GET("/admins", controller.AdminList)
			auth.POST("/admins", controller.SaveAdmin)
			auth.DELETE("/admins/:id", controller.DeleteAdmin)
			auth.GET("/logs", controller.OperationLogList)

			// 入驻审核
			auth.GET("/applies", controller.AdminApplyList)
			auth.PUT("/applies/:id/audit", controller.AdminAuditMerchant)
			auth.POST("/batch-audit-applies", controller.AdminBatchAuditMerchant)

			// 活动管理
			auth.GET("/activities", controller.AdminActivityList)
			auth.PUT("/activities/:id/status", controller.AdminChangeActivityStatus)
			auth.PUT("/activities/:id/audit", controller.AdminAuditActivity)
			auth.DELETE("/activities/:id", controller.AdminDeleteActivity)

			// 报名管理
			auth.GET("/signups", controller.AdminSignupList)
			auth.GET("/signups/:id", controller.AdminSignupDetail)
			auth.PUT("/signups/:id/audit", controller.AdminAuditSignup)
			auth.POST("/batch-audit-signups", controller.AdminBatchAuditSignups)
			auth.GET("/activities/:id/signups/export", controller.AdminExportSignups)

			// 用户管理
			auth.GET("/users", controller.AdminUserList)
			auth.PUT("/users/status", controller.AdminChangeUserStatus)
			auth.POST("/users/:id/revoke-merchant", controller.AdminRevokeMerchant)
		}
	}
}
