package myAuth

import (
	"context"
	"testing"
	"time"
)

func testJWTKey() string {
	return "0123456789abcdef0123456789abcdef"
}

func TestEncryptedJWTProviderRoundTrip(t *testing.T) {
	cfg := Config{
		Provider:           ProviderEncryptedJWT,
		TokenExpireSeconds: 3600,
		JWT: JWTConfig{
			Key:    testJWTKey(),
			Issuer: "test",
		},
	}
	p, err := newEncryptedJWTProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}

	claims := &Claims{
		UserID:      1001,
		Username:    "admin",
		DisplayName: "Admin",
		ExpireAt:    time.Now().Add(time.Hour),
	}
	token, err := p.Issue(context.Background(), claims)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	parsed, err := p.Parse(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.UserID != 1001 || parsed.Username != "admin" {
		t.Fatalf("unexpected claims: %+v", parsed)
	}
}

func TestInitWithEncryptedJWT(t *testing.T) {
	cfgMu.Lock()
	prevInit := initialized
	prevCfg := globalCfg
	prevMgr := globalManager
	cfgMu.Unlock()

	t.Cleanup(func() {
		cfgMu.Lock()
		initialized = prevInit
		globalCfg = prevCfg
		globalManager = prevMgr
		cfgMu.Unlock()
	})

	err := Init(Config{
		Provider:           ProviderEncryptedJWT,
		TokenExpireSeconds: 3600,
		JWT: JWTConfig{
			Key:    testJWTKey(),
			Issuer: "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	sess, token, err := Manager().Create(context.Background(), &SessionInput{
		UserID:      42,
		Username:    "u",
		DisplayName: "U",
	})
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || sess.UserID != 42 {
		t.Fatalf("unexpected session: token=%q sess=%+v", token, sess)
	}
}

func TestSessionExtras(t *testing.T) {
	s := &Session{UserID: 1}
	s.SetExtra("user.permSet", "ok")
	v, ok := s.Extra("user.permSet")
	if !ok || v != "ok" {
		t.Fatalf("extra not stored")
	}
}
