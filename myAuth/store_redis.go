package myAuth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure"
)

type sessionRecord struct {
	UserID      int64          `json:"userId"`
	Username    string         `json:"username"`
	DisplayName string         `json:"displayName"`
	ExpireAt    time.Time      `json:"expireAt"`
	JTI         string         `json:"jti"`
	Extras      map[string]any `json:"extras,omitempty"`
}

type redisSessionStore struct {
	prefix string
	ttl    time.Duration
}

func newRedisSessionStore(prefix string, ttl time.Duration) *redisSessionStore {
	return &redisSessionStore{prefix: prefix, ttl: ttl}
}

func (s *redisSessionStore) Save(ctx context.Context, token string, sess *Session) error {
	client := infrastructure.GetRedis()
	if client == nil {
		return fmt.Errorf("myAuth: redis is not initialized")
	}
	record := sessionRecord{
		UserID:      sess.UserID,
		Username:    sess.Username,
		DisplayName: sess.DisplayName,
		ExpireAt:    sess.ExpireAt,
		JTI:         sess.JTI,
		Extras:      sess.cloneExtras(),
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ttl := s.ttl
	if !sess.ExpireAt.IsZero() {
		if remain := time.Until(sess.ExpireAt); remain > 0 {
			ttl = remain
		}
	}
	return client.Set(ctx, s.prefix+token, data, ttl).Err()
}

func (s *redisSessionStore) Get(ctx context.Context, token string) (*Session, error) {
	client := infrastructure.GetRedis()
	if client == nil {
		return nil, fmt.Errorf("myAuth: redis is not initialized")
	}
	data, err := client.Get(ctx, s.prefix+token).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record sessionRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	sess := &Session{
		UserID:      record.UserID,
		Username:    record.Username,
		DisplayName: record.DisplayName,
		ExpireAt:    record.ExpireAt,
		JTI:         record.JTI,
		Token:       token,
	}
	if len(record.Extras) > 0 {
		sess.extras = record.Extras
	}
	return sess, nil
}

func (s *redisSessionStore) Delete(ctx context.Context, token string) error {
	client := infrastructure.GetRedis()
	if client == nil {
		return fmt.Errorf("myAuth: redis is not initialized")
	}
	return client.Del(ctx, s.prefix+token).Err()
}

type redisRevocationStore struct {
	prefix string
}

func newRedisRevocationStore(prefix string) *redisRevocationStore {
	return &redisRevocationStore{prefix: prefix}
}

func (s *redisRevocationStore) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	client := infrastructure.GetRedis()
	if client == nil {
		return false, nil
	}
	n, err := client.Exists(ctx, s.prefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *redisRevocationStore) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" {
		return nil
	}
	client := infrastructure.GetRedis()
	if client == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return client.Set(ctx, s.prefix+jti, "1", ttl).Err()
}
