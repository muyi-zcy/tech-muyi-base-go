package myLocale

import (
	"context"

	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
)

// Initialized 是否已加载文案
func Initialized() bool {
	return defaultStore != nil
}

// DefaultLocale 返回默认语言
func DefaultLocale() string {
	if defaultStore == nil {
		return "zh-CN"
	}
	return defaultStore.defaultLocale
}

// ResolveFromContext 从 context 读取 locale 并解析文案
func ResolveFromContext(ctx context.Context, code string, args map[string]string) string {
	locale := myContext.GetLocale(ctx)
	return Resolve(code, locale, args)
}

// HTTPHint 返回错误码对应的 HTTP 语义码，未知时返回 0
func HTTPHint(code string) int {
	if defaultStore == nil {
		return 0
	}
	return defaultStore.httpHint(code)
}

func (s *store) httpHint(code string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hint, ok := s.httpHints[code]; ok {
		return hint
	}
	return 0
}

func mergeMessages(base, overlay map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(overlay))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range overlay {
		merged[k] = v
	}
	return merged
}

func mergeHTTPHints(base, overlay map[string]int) map[string]int {
	merged := make(map[string]int, len(base)+len(overlay))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range overlay {
		merged[k] = v
	}
	return merged
}

func buildHTTPHints(contract *ErrorContract) map[string]int {
	hints := make(map[string]int, len(contract.Codes))
	for _, item := range contract.Codes {
		if item.HTTPHint > 0 {
			hints[item.Code] = item.HTTPHint
		}
	}
	return hints
}
