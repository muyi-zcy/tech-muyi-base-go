package myAuth

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
)

const (
	ProviderEncryptedJWT = "encrypted_jwt"
	ProviderRedisOpaque  = "redis_opaque"

	defaultTokenExpireSeconds = 28800
	defaultSessionKeyPrefix   = "auth:session:"
	defaultRevokeKeyPrefix    = "auth:revoked:"
)

// Config myAuth 配置。
type Config struct {
	Provider           string   `mapstructure:"provider"`
	TokenExpireSeconds int      `mapstructure:"tokenExpireSeconds"`
	WhiteList          []string `mapstructure:"whiteList"`

	JWT     JWTConfig          `mapstructure:"jwt"`
	Session SessionStoreConfig `mapstructure:"session"`
}

// JWTConfig 加密 JWT（JWE）配置。
type JWTConfig struct {
	Key      string `mapstructure:"key"`
	KeyFile  string `mapstructure:"keyFile"`
	Issuer   string `mapstructure:"issuer"`
	Audience string `mapstructure:"audience"`
}

// SessionStoreConfig Redis 会话存储配置（redis_opaque 及吊销）。
type SessionStoreConfig struct {
	KeyPrefix    string `mapstructure:"keyPrefix"`
	RevokePrefix string `mapstructure:"revokePrefix"`
}

var (
	cfgMu       sync.RWMutex
	globalCfg   Config
	initialized bool
)

// Init 初始化 myAuth，应在路由注册前调用。
func Init(cfg Config) error {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return err
	}

	manager, err := newManager(cfg)
	if err != nil {
		return err
	}

	cfgMu.Lock()
	defer cfgMu.Unlock()
	globalCfg = cfg
	globalManager = manager
	initialized = true
	return nil
}

// ConfigFromViper 从配置文件加载 [auth]（含 auth.jwt / auth.session 子段）。
func ConfigFromViper() (Config, error) {
	cfg := Config{}
	if err := config.GetConfigByType("auth", &cfg); err != nil {
		return Config{}, err
	}
	return cfg.withDefaults(), nil
}

// MustInitFromViper 从配置初始化，失败 panic。
func MustInitFromViper() {
	cfg, err := ConfigFromViper()
	if err != nil {
		panic(err)
	}
	if err := Init(cfg); err != nil {
		panic(err)
	}
}

// CurrentConfig 返回当前配置副本。
func CurrentConfig() Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return globalCfg
}

func (c Config) withDefaults() Config {
	if c.Provider == "" {
		c.Provider = ProviderEncryptedJWT
	}
	if c.TokenExpireSeconds <= 0 {
		c.TokenExpireSeconds = defaultTokenExpireSeconds
	}
	if c.Session.KeyPrefix == "" {
		c.Session.KeyPrefix = defaultSessionKeyPrefix
	}
	if c.Session.RevokePrefix == "" {
		c.Session.RevokePrefix = defaultRevokeKeyPrefix
	}
	if c.JWT.Issuer == "" {
		c.JWT.Issuer = "my-xi"
	}
	return c
}

func (c Config) validate() error {
	switch c.Provider {
	case ProviderEncryptedJWT:
		if _, err := c.JWT.decodeKey(); err != nil {
			return fmt.Errorf("myAuth: invalid auth.jwt.key: %w", err)
		}
	case ProviderRedisOpaque:
		// Redis 在运行时由 infrastructure 提供。
	default:
		providerMu.RLock()
		_, ok := providers[c.Provider]
		providerMu.RUnlock()
		if !ok {
			return fmt.Errorf("myAuth: unknown auth provider %q", c.Provider)
		}
	}
	return nil
}

func (c Config) tokenTTL() time.Duration {
	return time.Duration(c.TokenExpireSeconds) * time.Second
}

func (c JWTConfig) decodeKey() ([]byte, error) {
	raw := strings.TrimSpace(c.Key)
	if raw == "" {
		return nil, fmt.Errorf("jwt key is empty")
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	return nil, fmt.Errorf("jwt key must be 32 bytes or base64-encoded 32 bytes")
}

func ensureInitialized() {
	cfgMu.RLock()
	ok := initialized
	cfgMu.RUnlock()
	if !ok {
		panic("myAuth: call myAuth.Init or myAuth.MustInitFromViper before use")
	}
}
