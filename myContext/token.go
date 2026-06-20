package myContext

import (
	"context"

	"github.com/gin-gonic/gin"
)

func tokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(keyToken).(string); ok {
		return token
	}
	return ""
}

// resolveTokenFromGin 仅从 header/cookie 读取 token，供 Ingress 预置（ssoId 不可信，不读取）。
func resolveTokenFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, exists := c.Get(ginKeyToken); exists {
		if token, ok := v.(string); ok && token != "" {
			return token
		}
	}
	if token := c.GetHeader(HeaderToken); token != "" {
		return token
	}
	if token, _ := c.Cookie(HeaderToken); token != "" {
		return token
	}
	return ""
}

// TryGetToken 从 context 获取 token。
func TryGetToken(ctx context.Context) string {
	return tokenFromContext(ctx)
}

// WithToken 写入 token。
func WithToken(ctx context.Context, token string) context.Context {
	if ctx == nil || token == "" {
		return ctx
	}
	return context.WithValue(ctx, keyToken, token)
}
