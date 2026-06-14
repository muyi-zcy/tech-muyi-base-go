package myLocale

import (
	"strings"
)

// Resolve 按 locale 解析错误码文案，支持 {{arg}} 占位符
func Resolve(code, locale string, args map[string]string) string {
	if defaultStore == nil {
		return code
	}
	return defaultStore.resolve(code, locale, args)
}

func (s *store) resolve(code, locale string, args map[string]string) string {
	loc := locale
	if loc == "" {
		loc = s.defaultLocale
	}

	template := s.lookupMessage(code, loc)
	if template == "" {
		template = s.lookupMessage(code, s.defaultLocale)
	}
	if template == "" {
		return code
	}

	result := template
	for k, v := range args {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

func (s *store) lookupMessage(code, locale string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if bundle, ok := s.messages[locale]; ok {
		if msg, ok := bundle[code]; ok {
			return msg
		}
	}
	return ""
}
