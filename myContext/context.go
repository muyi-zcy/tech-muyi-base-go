package myContext

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
)

const (
	TraceId    = "traceId"
	SsoId      = "ssoId"
	SsoVersion = "ssoVersion"
)

// GetTraceId 从Gin上下文获取traceId
func GetTraceIdFromGinCtx(c *gin.Context) string {
	// 首先从Gin上下文获取
	if traceId, exists := c.Get(string(TraceId)); exists {
		if id, ok := traceId.(string); ok {
			return id
		}
	}

	// 从请求头获取
	traceId := c.GetHeader("X-Trace-ID")
	if traceId != "" {
		c.Set(string(TraceId), traceId)
		return traceId
	}

	// 从Cookie获取
	traceId, _ = c.Cookie("X-Trace-ID")
	if traceId != "" {
		c.Set(string(TraceId), traceId)
		return traceId
	}

	// 都没有则生成新的UUID
	newTraceId := uuid.New().String()
	c.Set(string(TraceId), newTraceId)
	c.Header("X-Trace-ID", newTraceId) // 设置响应头
	return newTraceId
}

func GetSsoIdFromGinCtx(c *gin.Context) string {
	// 首先从Gin上下文获取
	if ssoId, exists := c.Get(string(SsoId)); exists {
		if id, ok := ssoId.(string); ok {
			return id
		}
	}

	// 从请求头获取
	ssoId := c.GetHeader("X-Sso-ID")
	if ssoId != "" {
		c.Set(string(SsoId), ssoId)
		return ssoId
	}

	// 从Cookie获取
	ssoId, _ = c.Cookie("X-Sso-ID")
	if ssoId != "" {
		c.Set(string(SsoId), ssoId)
		return ssoId
	}

	// 都没有则生成新的UUID
	newSsoId := "-1"
	c.Set(string(SsoId), newSsoId)
	c.Header("X-Sso-ID", newSsoId) // 设置响应头
	return newSsoId
}

// ContextMiddleware 上下文中间件，用于初始化请求上下文
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取或生成traceId
		traceId := GetTraceIdFromGinCtx(c)

		// 将Gin上下文转换为标准context.Context并添加traceId
		// 使用自定义类型作为key，避免字符串冲突
		ctx := context.WithValue(c.Request.Context(), TraceId, traceId)
		c.Request = c.Request.WithContext(ctx)

		ssoId := GetSsoIdFromGinCtx(c)
		ctx = context.WithValue(c.Request.Context(), SsoId, ssoId)
		c.Request = c.Request.WithContext(ctx)

		// 继续处理请求
		c.Next()
	}
}

// FromContext 从标准context中获取traceId
func traceIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 使用自定义类型作为key，避免字符串冲突
	if traceId, ok := ctx.Value(TraceId).(string); ok {
		return traceId
	}
	return ""
}
func ssoIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 使用自定义类型作为key，避免字符串冲突
	if ssoId, ok := ctx.Value(SsoId).(string); ok {
		return ssoId
	}
	return ""
}

// GetTraceIdFromContext 从context中获取traceId的便捷方法
func GetTraceId(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", myException.NewExceptionFromError(myException.UNAUTHORIZED)
	}

	result := traceIdFromContext(ctx)
	if result == "" {
		return "", myException.NewExceptionFromError(myException.UNAUTHORIZED)
	}

	return result, nil
}

// GetSsoId 获取ssoId并进行非空验证，返回ssoId和异常
func GetSsoId(ctx context.Context) (string, error) {
	// 非空判断
	if ctx == nil {
		return "", myException.NewExceptionFromError(myException.UNAUTHORIZED)
	}

	result := ssoIdFromContext(ctx)
	if result == "" {
		return "", myException.NewExceptionFromError(myException.UNAUTHORIZED)
	}

	return result, nil
}
