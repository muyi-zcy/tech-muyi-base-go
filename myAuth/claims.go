package myAuth

import "time"

// Claims Token 解析后的标准载荷。
type Claims struct {
	UserID      int64
	Username    string
	DisplayName string
	ExpireAt    time.Time
	JTI         string
}

func claimsToSession(claims *Claims) *Session {
	if claims == nil {
		return nil
	}
	return &Session{
		UserID:      claims.UserID,
		Username:    claims.Username,
		DisplayName: claims.DisplayName,
		ExpireAt:    claims.ExpireAt,
		JTI:         claims.JTI,
	}
}
