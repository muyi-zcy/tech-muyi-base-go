package myContext

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ScalarBinding 可传播的标量上下文；空字符串表示不写入/不覆盖。
type ScalarBinding struct {
	SsoId string
	Token string
}

// BindScalars 将标量写入 Gin 与 request context（唯一入口，避免双写遗漏）。
func BindScalars(c *gin.Context, b ScalarBinding) {
	if c == nil || c.Request == nil {
		return
	}
	ctx := c.Request.Context()
	if b.SsoId != "" {
		c.Set(ginKeySso, b.SsoId)
		ctx = context.WithValue(ctx, keySso, b.SsoId)
	}
	if b.Token != "" {
		c.Set(ginKeyToken, b.Token)
		ctx = context.WithValue(ctx, keyToken, b.Token)
	}
	c.Request = c.Request.WithContext(ctx)
}

// BindTrace 写入 traceId（HTTP Ingress 专用）。
func BindTrace(c *gin.Context, traceId string) {
	if c == nil || c.Request == nil || traceId == "" {
		return
	}
	c.Set(ginKeyTrace, traceId)
	ctx := context.WithValue(c.Request.Context(), keyTrace, traceId)
	c.Request = c.Request.WithContext(ctx)
}

// AttachContext 替换 request context（Session 等对象域使用）。
func AttachContext(c *gin.Context, ctx context.Context) {
	if c == nil || ctx == nil {
		return
	}
	c.Request = c.Request.WithContext(ctx)
}
