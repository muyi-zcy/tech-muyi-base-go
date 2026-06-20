package myAuth

import "strings"

// WhiteListMatcher 白名单匹配器，支持精确路径与通配符 * / **。
type WhiteListMatcher struct {
	exact    map[string]struct{}
	patterns []string
}

// NewWhiteListMatcher 创建白名单匹配器。
func NewWhiteListMatcher(entries []string) *WhiteListMatcher {
	m := &WhiteListMatcher{
		exact: make(map[string]struct{}),
	}
	for _, entry := range entries {
		entry = NormalizePath(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "*") {
			m.patterns = append(m.patterns, entry)
			continue
		}
		m.exact[entry] = struct{}{}
	}
	return m
}

// Match 判断 path 是否命中白名单。
func (m *WhiteListMatcher) Match(path string) bool {
	if m == nil {
		return false
	}
	path = NormalizePath(path)
	if path == "" {
		return false
	}
	if _, ok := m.exact[path]; ok {
		return true
	}
	for _, pattern := range m.patterns {
		if matchPathPattern(pattern, path) {
			return true
		}
	}
	return false
}

// IsWhiteListed 便捷函数。
func IsWhiteListed(path string, entries []string) bool {
	return NewWhiteListMatcher(entries).Match(path)
}

// NormalizePath 规范化 URL 路径。
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func matchPathPattern(pattern, path string) bool {
	pattern = NormalizePath(pattern)
	path = NormalizePath(path)
	if pattern == path {
		return true
	}

	pSegs := splitPath(pattern)
	uSegs := splitPath(path)
	return matchSegments(pSegs, uSegs)
}

func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func matchSegments(pattern, path []string) bool {
	pi, ui := 0, 0
	for pi < len(pattern) {
		if pattern[pi] == "**" {
			if pi == len(pattern)-1 {
				return true
			}
			for ui <= len(path) {
				if matchSegments(pattern[pi+1:], path[ui:]) {
					return true
				}
				ui++
			}
			return false
		}
		if ui >= len(path) {
			return false
		}
		if pattern[pi] != "*" && pattern[pi] != path[ui] {
			return false
		}
		pi++
		ui++
	}
	return ui == len(path)
}
