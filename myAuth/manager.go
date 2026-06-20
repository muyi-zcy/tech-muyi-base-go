package myAuth

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionManager 会话管理入口。
type SessionManager interface {
	Create(ctx context.Context, input *SessionInput) (*Session, string, error)
	LoadFromRequest(c *gin.Context, providerName string) (*Session, error)
	LoadFromToken(ctx context.Context, token string, providerName string) (*Session, error)
	Destroy(ctx context.Context, token string) error
}

var globalManager SessionManager

// Manager 返回全局 SessionManager。
func Manager() SessionManager {
	ensureInitialized()
	return globalManager
}

type manager struct {
	cfg      Config
	provider TokenProvider
	matcher  *WhiteListMatcher
}

func newManager(cfg Config) (*manager, error) {
	provider, err := resolveProvider(cfg.Provider, cfg)
	if err != nil {
		return nil, err
	}
	RegisterTokenProvider(provider)
	return &manager{
		cfg:      cfg,
		provider: provider,
		matcher:  NewWhiteListMatcher(cfg.WhiteList),
	}, nil
}

func (m *manager) Create(ctx context.Context, input *SessionInput) (*Session, string, error) {
	if input == nil {
		return nil, "", fmt.Errorf("myAuth: session input is nil")
	}
	ttl := m.cfg.tokenTTL()
	if input.TTL > 0 {
		ttl = input.TTL
	}
	claims := &Claims{
		UserID:      input.UserID,
		Username:    input.Username,
		DisplayName: input.DisplayName,
		ExpireAt:    time.Now().Add(ttl),
	}
	token, err := m.provider.Issue(ctx, claims)
	if err != nil {
		return nil, "", err
	}
	sess := claimsToSession(claims)
	sess.Token = token
	sess.JTI = claims.JTI
	return sess, token, nil
}

func (m *manager) LoadFromRequest(c *gin.Context, providerName string) (*Session, error) {
	token := ExtractToken(c)
	if token == "" {
		return nil, nil
	}
	return m.LoadFromToken(c.Request.Context(), token, providerName)
}

func (m *manager) LoadFromToken(ctx context.Context, token string, providerName string) (*Session, error) {
	if token == "" {
		return nil, nil
	}

	provider := m.provider
	if providerName != "" && providerName != m.cfg.Provider {
		p, err := resolveProvider(providerName, m.cfg)
		if err != nil {
			return nil, err
		}
		provider = p
	}

	claims, err := provider.Parse(ctx, token)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, nil
	}
	sess := claimsToSession(claims)
	sess.Token = token
	if sess.JTI == "" {
		sess.JTI = claims.JTI
	}
	return sess, nil
}

func (m *manager) Destroy(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	claims, err := m.provider.Parse(ctx, token)
	if err != nil {
		return err
	}
	if claims == nil {
		return nil
	}
	return m.provider.Revoke(ctx, claims)
}

func (m *manager) whiteListMatch(path string) bool {
	return m.matcher.Match(path)
}
