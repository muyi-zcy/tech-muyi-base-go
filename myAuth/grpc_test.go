package myAuth

import (
	"context"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestGRPCAuthInterceptorLoadsSession(t *testing.T) {
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

	if err := Init(Config{
		Provider:           ProviderEncryptedJWT,
		TokenExpireSeconds: 3600,
		JWT: JWTConfig{
			Key:    testJWTKey(),
			Issuer: "test",
		},
	}); err != nil {
		t.Fatal(err)
	}

	sess, token, err := Manager().Create(context.Background(), &SessionInput{
		UserID:      99,
		Username:    "grpc-user",
		DisplayName: "GRPC User",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil || token == "" {
		t.Fatal("empty session or token")
	}

	interceptor := GRPCAuthInterceptor()
	handler := func(ctx context.Context, req any) (any, error) {
		got, ok := SessionFromContext(ctx)
		if !ok || got == nil {
			t.Fatal("session missing in gRPC context")
		}
		if got.UserID != 99 {
			t.Fatalf("userID = %d", got.UserID)
		}
		if myContext.TryGetSsoId(ctx) != "99" {
			t.Fatalf("ssoId = %q", myContext.TryGetSsoId(ctx))
		}
		return "ok", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		myContext.HeaderToken, token,
	))
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Test"}
	_, err = interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatal(err)
	}
}
