package myAuth

import (
	"context"
	"sync"
)

// TokenProvider Token 签发与解析。
type TokenProvider interface {
	Name() string
	Issue(ctx context.Context, claims *Claims) (token string, err error)
	Parse(ctx context.Context, token string) (*Claims, error)
	Revoke(ctx context.Context, claims *Claims) error
}

var (
	providerMu sync.RWMutex
	providers  = make(map[string]TokenProvider)
)

// RegisterTokenProvider 注册自定义 TokenProvider。
func RegisterTokenProvider(p TokenProvider) {
	if p == nil {
		return
	}
	providerMu.Lock()
	defer providerMu.Unlock()
	providers[p.Name()] = p
}

func getProvider(name string) (TokenProvider, bool) {
	providerMu.RLock()
	defer providerMu.RUnlock()
	p, ok := providers[name]
	return p, ok
}

func resolveProvider(name string, cfg Config) (TokenProvider, error) {
	if p, ok := getProvider(name); ok {
		return p, nil
	}
	switch name {
	case ProviderEncryptedJWT:
		return newEncryptedJWTProvider(cfg)
	case ProviderRedisOpaque:
		return newRedisOpaqueProvider(cfg)
	default:
		return nil, errUnknownProvider(name)
	}
}

type providerError string

func (e providerError) Error() string { return string(e) }

func errUnknownProvider(name string) error {
	return providerError("myAuth: unknown token provider " + name)
}
