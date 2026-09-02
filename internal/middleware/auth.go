// Package middleware 提供跨域、日志、限流、鉴权等通用中间件。
package middleware

import (
	"strings"

	"activity/pkg/errcode"
	"activity/pkg/jwtx"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserAuth 小程序用户鉴权：校验 JWT 并注入用户 ID 与角色到上下文
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Abort(c, errcode.ErrUnauth)
			return
		}
		claims, err := jwtx.ParseUserToken(token)
		if err != nil {
			response.Abort(c, errcode.ErrUnauth)
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// OptionalUserAuth 可选鉴权：携带有效令牌时注入用户信息，否则以游客身份继续
func OptionalUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Set("uid", int64(0))
			c.Set("role", int8(0))
			c.Next()
			return
		}
		claims, err := jwtx.ParseUserToken(token)
		if err != nil {
			c.Set("uid", int64(0))
			c.Set("role", int8(0))
			c.Next()
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminAuth 管理端鉴权：校验 JWT 并注入管理员信息
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Abort(c, errcode.ErrUnauth)
			return
		}
		claims, err := jwtx.ParseAdminToken(token)
		if err != nil {
			response.Abort(c, errcode.ErrUnauth)
			return
		}
		c.Set("aid", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	raw := c.GetHeader("Authorization")
	if raw == "" {
		return ""
	}
	if len(raw) > 7 && strings.EqualFold(raw[:7], "Bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return strings.TrimSpace(raw)
}

// GetUserID 从上下文读取当前登录用户 ID
func GetUserID(c *gin.Context) int64 {
	v, _ := c.Get("uid")
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

// GetUserRole 从上下文读取当前登录用户角色
func GetUserRole(c *gin.Context) int8 {
	v, _ := c.Get("role")
	if role, ok := v.(int8); ok {
		return role
	}
	return 0
}

// GetAdminID 从上下文读取当前管理员 ID
func GetAdminID(c *gin.Context) int64 {
	v, _ := c.Get("aid")
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

// GetAdminRole 从上下文读取当前管理员角色
func GetAdminRole(c *gin.Context) int8 {
	v, _ := c.Get("admin_role")
	if role, ok := v.(int8); ok {
		return role
	}
	return 2
}

// GetAdminName 从上下文读取当前管理员账号
func GetAdminName(c *gin.Context) string {
	v, _ := c.Get("username")
	if name, ok := v.(string); ok {
		return name
	}
	return ""
}

// GetClientIP 获取客户端真实 IP
func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
