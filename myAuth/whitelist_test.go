package myAuth

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/api/user/v1/", "/api/user/v1"},
		{"/api//user/v1", "/api/user/v1"},
		{"/", "/"},
	}
	for _, tt := range tests {
		if got := NormalizePath(tt.in); got != tt.want {
			t.Fatalf("NormalizePath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestWhiteListMatcher(t *testing.T) {
	m := NewWhiteListMatcher([]string{
		"/api/user/v1/auth/login",
		"/api/user/v1/open/**",
		"/api/*/v1/health",
		"/api/open/*",
	})

	cases := []struct {
		path string
		want bool
	}{
		{"/api/user/v1/auth/login", true},
		{"/api/user/v1/open/register", true},
		{"/api/user/v1/open/a/b", true},
		{"/api/file/v1/health", true},
		{"/api/open/a", true},
		{"/api/open/a/b", false},
		{"/api/user/v1/user/save", false},
	}
	for _, c := range cases {
		if got := m.Match(c.path); got != c.want {
			t.Fatalf("Match(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestNormalizeToken(t *testing.T) {
	if got := NormalizeToken("  Bearer abc  "); got != "abc" {
		t.Fatalf("NormalizeToken = %q", got)
	}
}
