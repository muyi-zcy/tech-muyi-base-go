package myContext

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	LocaleKey = "locale"

	acceptLanguageHeader = "Accept-Language"
)

// LocaleMiddleware 解析 locale 并写入 Gin / request context
func LocaleMiddleware(defaultLocale string) gin.HandlerFunc {
	if defaultLocale == "" {
		defaultLocale = "zh-CN"
	}
	return func(c *gin.Context) {
		locale := resolveLocale(c, defaultLocale)
		c.Set(string(LocaleKey), locale)
		ctx := context.WithValue(c.Request.Context(), LocaleKey, locale)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func resolveLocale(c *gin.Context, defaultLocale string) string {
	if locale := strings.TrimSpace(c.Query("locale")); locale != "" {
		return normalizeLocale(locale)
	}
	if locale := parseAcceptLanguage(c.GetHeader(acceptLanguageHeader)); locale != "" {
		return locale
	}
	return defaultLocale
}

func parseAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}
	first := strings.TrimSpace(parts[0])
	if semi := strings.Index(first, ";"); semi >= 0 {
		first = strings.TrimSpace(first[:semi])
	}
	return normalizeLocale(first)
}

func normalizeLocale(locale string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return ""
	}
	segments := strings.Split(locale, "-")
	if len(segments) == 1 {
		return strings.ToLower(segments[0])
	}
	return strings.ToLower(segments[0]) + "-" + strings.ToUpper(segments[1])
}

// GetLocaleFromGinCtx 从 Gin 上下文获取 locale
func GetLocaleFromGinCtx(c *gin.Context) string {
	if locale, ok := c.Get(string(LocaleKey)); ok {
		if value, ok := locale.(string); ok && value != "" {
			return value
		}
	}
	return "zh-CN"
}

// GetLocale 从标准 context 获取 locale
func GetLocale(ctx context.Context) string {
	if ctx == nil {
		return "zh-CN"
	}
	if locale, ok := ctx.Value(LocaleKey).(string); ok && locale != "" {
		return locale
	}
	return "zh-CN"
}
