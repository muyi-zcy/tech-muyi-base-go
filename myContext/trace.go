package myContext

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
)

func traceIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceId, ok := ctx.Value(keyTrace).(string); ok {
		return traceId
	}
	return ""
}

// resolveTraceIdFromGin 解析 traceId：Gin 缓存 → header → cookie → 生成 UUID。
func resolveTraceIdFromGin(c *gin.Context) string {
	if c == nil {
		return uuid.New().String()
	}
	if traceId, exists := c.Get(ginKeyTrace); exists {
		if id, ok := traceId.(string); ok && id != "" {
			return id
		}
	}
	if traceId := c.GetHeader(HeaderTraceId); traceId != "" {
		return traceId
	}
	if traceId, _ := c.Cookie(HeaderTraceId); traceId != "" {
		return traceId
	}
	return uuid.New().String()
}

// TryGetTraceId 从 context 获取 traceId（不报错）。
func TryGetTraceId(ctx context.Context) string {
	return traceIdFromContext(ctx)
}

// GetTraceId 获取 traceId；缺失时返回 platform.unauthorized。
func GetTraceId(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", myException.NewBizError("platform.unauthorized", nil)
	}
	if result := traceIdFromContext(ctx); result != "" {
		return result, nil
	}
	return "", myException.NewBizError("platform.unauthorized", nil)
}
