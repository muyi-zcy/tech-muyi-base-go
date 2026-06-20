package myAuth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type encryptedJWTProvider struct {
	key      []byte
	issuer   string
	audience string
	ttl      time.Duration
	revoke   *redisRevocationStore
}

func newEncryptedJWTProvider(cfg Config) (TokenProvider, error) {
	key, err := cfg.JWT.decodeKey()
	if err != nil {
		return nil, err
	}
	return &encryptedJWTProvider{
		key:      key,
		issuer:   cfg.JWT.Issuer,
		audience: cfg.JWT.Audience,
		ttl:      cfg.tokenTTL(),
		revoke:   newRedisRevocationStore(cfg.Session.RevokePrefix),
	}, nil
}

func (p *encryptedJWTProvider) Name() string { return ProviderEncryptedJWT }

func (p *encryptedJWTProvider) Issue(ctx context.Context, claims *Claims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("myAuth: claims is nil")
	}
	now := time.Now()
	expireAt := claims.ExpireAt
	if expireAt.IsZero() {
		expireAt = now.Add(p.ttl)
	}
	jti := claims.JTI
	if jti == "" {
		jti = uuid.NewString()
	}
	claims.JTI = jti

	builder := jwt.NewBuilder().
		Issuer(p.issuer).
		Subject(strconv.FormatInt(claims.UserID, 10)).
		IssuedAt(now).
		Expiration(expireAt).
		JwtID(jti).
		Claim("username", claims.Username).
		Claim("displayName", claims.DisplayName)

	if p.audience != "" {
		builder = builder.Audience([]string{p.audience})
	}

	tok, err := builder.Build()
	if err != nil {
		return "", err
	}

	serialized, err := jwt.NewSerializer().
		Encrypt(
			jwt.WithKey(jwa.DIRECT, p.key),
			jwt.WithEncryptOption(jwe.WithContentEncryption(jwa.A256GCM)),
		).
		Serialize(tok)
	if err != nil {
		return "", err
	}
	return string(serialized), nil
}

func (p *encryptedJWTProvider) Parse(ctx context.Context, token string) (*Claims, error) {
	decrypted, err := jwe.Decrypt([]byte(token), jwe.WithKey(jwa.DIRECT, p.key))
	if err != nil {
		return nil, fmt.Errorf("myAuth: decrypt token: %w", err)
	}

	tok, err := jwt.ParseInsecure(decrypted)
	if err != nil {
		return nil, err
	}
	if err := jwt.Validate(tok); err != nil {
		return nil, err
	}

	jti, _ := tok.Get(jwt.JwtIDKey)
	jtiStr, _ := jti.(string)
	if revoked, err := p.revoke.IsRevoked(ctx, jtiStr); err != nil {
		return nil, err
	} else if revoked {
		return nil, fmt.Errorf("myAuth: token revoked")
	}

	sub, _ := tok.Get(jwt.SubjectKey)
	userID, _ := strconv.ParseInt(fmt.Sprint(sub), 10, 64)

	username, _ := tok.Get("username")
	displayName, _ := tok.Get("displayName")

	var expireAt time.Time
	if exp, ok := tok.Get(jwt.ExpirationKey); ok {
		switch v := exp.(type) {
		case time.Time:
			expireAt = v
		case float64:
			expireAt = time.Unix(int64(v), 0)
		}
	}

	return &Claims{
		UserID:      userID,
		Username:    fmt.Sprint(username),
		DisplayName: fmt.Sprint(displayName),
		ExpireAt:    expireAt,
		JTI:         jtiStr,
	}, nil
}

func (p *encryptedJWTProvider) Revoke(ctx context.Context, claims *Claims) error {
	if claims == nil {
		return nil
	}
	ttl := time.Until(claims.ExpireAt)
	return p.revoke.Revoke(ctx, claims.JTI, ttl)
}

type redisOpaqueProvider struct {
	store  *redisSessionStore
	revoke *redisRevocationStore
	ttl    time.Duration
}

func newRedisOpaqueProvider(cfg Config) (TokenProvider, error) {
	return &redisOpaqueProvider{
		store:  newRedisSessionStore(cfg.Session.KeyPrefix, cfg.tokenTTL()),
		revoke: newRedisRevocationStore(cfg.Session.RevokePrefix),
		ttl:    cfg.tokenTTL(),
	}, nil
}

func (p *redisOpaqueProvider) Name() string { return ProviderRedisOpaque }

func (p *redisOpaqueProvider) Issue(ctx context.Context, claims *Claims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("myAuth: claims is nil")
	}
	token := claims.JTI
	if token == "" {
		token = uuid.NewString()
		claims.JTI = token
	}
	if claims.ExpireAt.IsZero() {
		claims.ExpireAt = time.Now().Add(p.ttl)
	}
	sess := claimsToSession(claims)
	sess.Token = token
	if err := p.store.Save(ctx, token, sess); err != nil {
		return "", err
	}
	return token, nil
}

func (p *redisOpaqueProvider) Parse(ctx context.Context, token string) (*Claims, error) {
	sess, err := p.store.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	if !sess.ExpireAt.IsZero() && time.Now().After(sess.ExpireAt) {
		return nil, fmt.Errorf("myAuth: token expired")
	}
	jti := sess.JTI
	if jti == "" {
		jti = token
	}
	return &Claims{
		UserID:      sess.UserID,
		Username:    sess.Username,
		DisplayName: sess.DisplayName,
		ExpireAt:    sess.ExpireAt,
		JTI:         jti,
	}, nil
}

func (p *redisOpaqueProvider) Revoke(ctx context.Context, claims *Claims) error {
	if claims == nil || claims.JTI == "" {
		return nil
	}
	if err := p.store.Delete(ctx, claims.JTI); err != nil {
		return err
	}
	ttl := time.Until(claims.ExpireAt)
	return p.revoke.Revoke(ctx, claims.JTI, ttl)
}
