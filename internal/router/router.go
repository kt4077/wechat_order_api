package router

import (
	"strings"

	"activity/config"
	"activity/internal/controller"
	"activity/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

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

func registerCommon(g *gin.RouterGroup) {
	common := g.Group("/common")
	{
		common.POST("/login", controller.Login)
		common.GET("/config", controller.GetConfig)
	}
}

func registerUser(g *gin.RouterGroup) {
	public := g.Group("")
	public.Use(middleware.OptionalUserAuth())
	{
		public.GET("/activities", controller.ListActivities)
		public.GET("/activities/:id", controller.ActivityDetail)
	}

	auth := g.Group("")
	auth.Use(middleware.UserAuth())
	{
		auth.GET("/user/profile", controller.GetProfile)
		auth.PUT("/user/profile", controller.UpdateProfile)
		auth.DELETE("/user/data", controller.DeleteMyData)

		auth.GET("/messages", controller.ListMessages)
		auth.GET("/messages/unread", controller.UnreadCount)
		auth.PUT("/messages/read", controller.ReadMessage)
		auth.PUT("/messages/read-all", controller.ReadAllMessages)

		auth.POST("/common/upload", controller.UploadImage)

		auth.POST("/merchant/apply", controller.ApplyMerchant)
		auth.GET("/merchant/status", controller.MyMerchantStatus)

		auth.POST("/activities", controller.CreateActivity)
		auth.PUT("/activities/:id", controller.UpdateActivity)
		auth.DELETE("/activities/:id", controller.DeleteActivity)
		auth.POST("/activities/:id/copy", controller.CopyActivity)
		auth.PUT("/activities/:id/status", controller.ChangeActivityStatus)
		auth.GET("/activities/:id/stat", controller.ActivityStat)
		auth.GET("/activities/:id/signups", controller.ListSignups)
		auth.GET("/activities/:id/signups/export", controller.ExportSignups)

		auth.POST("/signups", controller.SubmitSignup)
		auth.GET("/signups", controller.ListSignups)
		auth.GET("/signups/:id", controller.SignupDetail)
		auth.POST("/signups/:id/cancel", controller.CancelSignup)
		auth.PUT("/signups/:id/audit", controller.AuditSignup)
		auth.POST("/batch-audit-signups", controller.BatchAuditSignups)
	}
}

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

			auth.GET("/config", controller.GetSysConfig)
			auth.PUT("/config", controller.SaveSysConfig)

			auth.GET("/admins", controller.AdminList)
			auth.POST("/admins", controller.SaveAdmin)
			auth.DELETE("/admins/:id", controller.DeleteAdmin)
			auth.GET("/logs", controller.OperationLogList)

			auth.GET("/applies", controller.AdminApplyList)
			auth.PUT("/applies/:id/audit", controller.AdminAuditMerchant)
			auth.POST("/batch-audit-applies", controller.AdminBatchAuditMerchant)

			auth.GET("/activities", controller.AdminActivityList)
			auth.PUT("/activities/:id/status", controller.AdminChangeActivityStatus)
			auth.PUT("/activities/:id/audit", controller.AdminAuditActivity)
			auth.DELETE("/activities/:id", controller.AdminDeleteActivity)

			auth.GET("/signups", controller.AdminSignupList)
			auth.GET("/signups/:id", controller.AdminSignupDetail)
			auth.PUT("/signups/:id/audit", controller.AdminAuditSignup)
			auth.POST("/batch-audit-signups", controller.AdminBatchAuditSignups)
			auth.GET("/activities/:id/signups/export", controller.AdminExportSignups)

			auth.GET("/users", controller.AdminUserList)
			auth.PUT("/users/status", controller.AdminChangeUserStatus)
			auth.POST("/users/:id/revoke-merchant", controller.AdminRevokeMerchant)
		}
	}
}
