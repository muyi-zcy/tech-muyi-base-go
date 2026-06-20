package myAuth

import "time"

// Session 平台标准会话，仅含身份字段；业务数据通过 extras 扩展。
type Session struct {
	UserID      int64
	Username    string
	DisplayName string
	ExpireAt    time.Time

	Token string
	JTI   string

	extras map[string]any
}

// SessionInput 创建会话时的输入。
type SessionInput struct {
	UserID      int64
	Username    string
	DisplayName string
	TTL         time.Duration
}

func (s *Session) Extra(key string) (any, bool) {
	if s == nil || s.extras == nil {
		return nil, false
	}
	v, ok := s.extras[key]
	return v, ok
}

func (s *Session) SetExtra(key string, val any) {
	if s == nil {
		return
	}
	if s.extras == nil {
		s.extras = make(map[string]any)
	}
	s.extras[key] = val
}

func (s *Session) cloneExtras() map[string]any {
	if s == nil || len(s.extras) == 0 {
		return nil
	}
	out := make(map[string]any, len(s.extras))
	for k, v := range s.extras {
		out[k] = v
	}
	return out
}
