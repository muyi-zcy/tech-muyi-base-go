package middleware

import (
	"bytes"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // 保存一份
	return w.ResponseWriter.Write(b)
}

// Logger 日志中间件 - 统一记录请求开始和结束日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		remoteIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 记录请求开始日志（自动获取traceId）
		myLogger.InfoCtx(c.Request.Context(), "Request started",
			zap.String("httpMethod", method),
			zap.String("httpPath", path),
			zap.String("remoteIp", remoteIP),
			zap.String("userAgent", userAgent),
		)

		// 处理请求
		c.Next()

		// 记录请求结束日志（自动获取traceId）
		end := time.Now()
		duration := end.Sub(start)
		statusCode := c.Writer.Status()
		responseBody := blw.body.String()

		// 合并请求开始和结束信息
		myLogger.InfoCtx(c.Request.Context(), "Request finished",
			zap.String("httpMethod", method),
			zap.String("httpPath", path),
			zap.String("remoteIp", remoteIP),
			zap.String("userAgent", userAgent),
			zap.Int("httpStatus", statusCode),
			zap.String("result", responseBody),
			zap.Int64("durationMs", duration.Milliseconds()),
			zap.Int64("durationMicroseconds", duration.Microseconds()),
		)
	}
}
