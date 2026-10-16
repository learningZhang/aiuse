package router

import (
	"net/http"
	"time"

	"mylog/provider"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestID 请求 ID 中间件：复用请求头中的 X-Request-Id，缺失则生成 UUID。
// ID 同时写入 context（供日志与响应体使用）和响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := provider.NewRequestID(c.GetHeader(provider.RequestIDHeader))
		ctx := provider.WithRequestID(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(provider.RequestIDHeader, id)
		c.Next()
	}
}

// Logger 访问日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		fields := []zap.Field{
			zapString("request_id", provider.RequestID(c.Request.Context())),
			zapString("method", c.Request.Method),
			zapString("path", c.Request.URL.Path),
			zapString("query", c.Request.URL.RawQuery),
			zapString("client_ip", c.ClientIP()),
			zapInt("status", c.Writer.Status()),
			zapInt64("latency_ms", time.Since(start).Milliseconds()),
		}
		if err := c.Errors.Last(); err != nil {
			fields = append(fields, provider.ErrField(err))
		}
		provider.Log.Info("http request", fields...)
	}
}

// Recovery panic 恢复中间件：记录堆栈并返回统一 500
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				provider.Log.Error("panic recovered",
					zapAny("panic", r),
					zapString("path", c.Request.URL.Path),
					zapString("request_id", provider.RequestID(c.Request.Context())),
					zapString("stack", stack(4<<10)),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "InternalError",
						"message":    "server inner error",
						"request_id": provider.RequestID(c.Request.Context()),
					},
				})
			}
		}()
		c.Next()
	}
}

// CORS 跨域中间件。allowed 为配置的来源白名单
func CORS(allowed []string) gin.HandlerFunc {
	set := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		set[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if set["*"] || (origin != "" && set[origin]) {
			c.Header("Access-Control-Allow-Origin", originOrAny(origin, set["*"]))
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,X-Request-Id")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originOrAny(origin string, anyOrigin bool) string {
	if anyOrigin && origin == "" {
		return "*"
	}
	return origin
}

func stack(size int) string {
	buf := make([]byte, size)
	n := runtimeStack(buf, false)
	return string(buf[:n])
}
