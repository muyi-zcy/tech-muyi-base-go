package myContext

import (
	"github.com/gin-gonic/gin"
)

// ContextMiddleware HTTP 入口中间件：初始化 traceId，可选预置 token；不读取/不信任 x-sso-id。
func ContextMiddleware() gin.HandlerFunc {
	return HTTPIngressMiddleware()
}

// HTTPIngressMiddleware 同 ContextMiddleware。
func HTTPIngressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := resolveTraceIdFromGin(c)
		BindTrace(c, traceId)
		c.Header(HeaderTraceId, traceId)

		if token := resolveTokenFromGin(c); token != "" {
			BindScalars(c, ScalarBinding{Token: token})
		}

		c.Next()
	}
}
