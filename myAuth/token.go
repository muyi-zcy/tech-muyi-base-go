package myAuth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
)

// ExtractToken 从请求中提取 token（Header x-token → Cookie x-token）。
func ExtractToken(c *gin.Context) string {
	return extractTokenFromRequest(c, myContext.HeaderToken)
}

func extractTokenFromRequest(c *gin.Context, headerName string) string {
	if c == nil || c.Request == nil {
		return ""
	}
	token := NormalizeToken(c.GetHeader(headerName))
	if token != "" {
		return token
	}
	if cookie, err := c.Cookie(headerName); err == nil {
		return NormalizeToken(cookie)
	}
	return ""
}

// NormalizeToken 规范化 token 字符串。
func NormalizeToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 7 && strings.EqualFold(raw[:7], "bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return raw
}

// Token 返回当前请求的 token；已鉴权时从 context 读取，否则从 Header/Cookie 提取。
func Token(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if t := myContext.TryGetToken(c.Request.Context()); t != "" {
		return t
	}
	return ExtractToken(c)
}
