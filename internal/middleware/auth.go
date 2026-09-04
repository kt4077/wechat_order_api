package middleware

import (
	"strings"

	"activity/pkg/errcode"
	"activity/pkg/jwtx"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

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

func GetUserID(c *gin.Context) int64 {
	v, _ := c.Get("uid")
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

func GetUserRole(c *gin.Context) int8 {
	v, _ := c.Get("role")
	if role, ok := v.(int8); ok {
		return role
	}
	return 0
}

func GetAdminID(c *gin.Context) int64 {
	v, _ := c.Get("aid")
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

func GetAdminRole(c *gin.Context) int8 {
	v, _ := c.Get("admin_role")
	if role, ok := v.(int8); ok {
		return role
	}
	return 2
}

func GetAdminName(c *gin.Context) string {
	v, _ := c.Get("username")
	if name, ok := v.(string); ok {
		return name
	}
	return ""
}

func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
