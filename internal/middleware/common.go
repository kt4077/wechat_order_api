package middleware

import (
	"net/http"
	"sync"
	"time"

	"activity/config"
	"activity/pkg/errcode"
	"activity/pkg/response"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	origins := config.Get().CORS.AllowOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allow := ""
		for _, item := range origins {
			if item == "*" {
				allow = "*"
				break
			}
			if item == origin {
				allow = origin
				break
			}
		}
		if allow != "" {
			c.Header("Access-Control-Allow-Origin", allow)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization,Accept,X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Disposition,Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type visitor struct {
	tokens   float64
	lastTime time.Time
}

var (
	visitors = map[string]*visitor{}
	limitMu  sync.Mutex
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get().RateLimit
		if !cfg.Enable || cfg.QPS <= 0 {
			c.Next()
			return
		}
		ip := GetClientIP(c)
		limitMu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{tokens: float64(cfg.Burst), lastTime: time.Now()}
			visitors[ip] = v
		}
		now := time.Now()
		elapsed := now.Sub(v.lastTime).Seconds()
		v.lastTime = now
		v.tokens += elapsed * float64(cfg.QPS)
		if v.tokens > float64(cfg.Burst) {
			v.tokens = float64(cfg.Burst)
		}
		if v.tokens < 1 {
			limitMu.Unlock()
			response.Abort(c, errcode.ErrTooMany)
			return
		}
		v.tokens--
		if len(visitors) > 10000 {
			for key, item := range visitors {
				if time.Since(item.lastTime) > 10*time.Minute {
					delete(visitors, key)
				}
			}
		}
		limitMu.Unlock()
		c.Next()
	}
}
