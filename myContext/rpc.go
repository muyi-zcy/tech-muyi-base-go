package myContext

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	Token               = "token"
	SourceService       = "sourceService"
	HeaderTraceId       = "x-trace-id"
	HeaderSsoId         = "x-sso-id"
	HeaderToken         = "x-token"
	HeaderSourceService = "x-source-service"
)

// TryGetTraceId 从 context 获取 traceId（不报错）
func TryGetTraceId(ctx context.Context) string {
	return traceIdFromContext(ctx)
}

// TryGetSsoId 从 context 获取 ssoId（不报错）
func TryGetSsoId(ctx context.Context) string {
	return ssoIdFromContext(ctx)
}

// TryGetToken 从 context 获取 token
func TryGetToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(Token).(string); ok {
		return token
	}
	return ""
}

// WithToken 写入 token
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, Token, token)
}

// WithMetadata 从 gRPC metadata 恢复上下文
func WithMetadata(ctx context.Context, md metadata.MD) context.Context {
	if v := firstMD(md, HeaderTraceId); v != "" {
		ctx = context.WithValue(ctx, TraceId, v)
	}
	if v := firstMD(md, HeaderSsoId); v != "" {
		ctx = context.WithValue(ctx, SsoId, v)
	}
	if v := firstMD(md, HeaderToken); v != "" {
		ctx = context.WithValue(ctx, Token, v)
	}
	if v := firstMD(md, HeaderSourceService); v != "" {
		ctx = context.WithValue(ctx, SourceService, v)
	}
	return ctx
}

// MetadataFromContext 从 context 提取 metadata 键值对
func MetadataFromContext(ctx context.Context) metadata.MD {
	md := metadata.MD{}
	if traceId := TryGetTraceId(ctx); traceId != "" {
		md.Set(HeaderTraceId, traceId)
	}
	if ssoId := TryGetSsoId(ctx); ssoId != "" {
		md.Set(HeaderSsoId, ssoId)
	}
	if token := TryGetToken(ctx); token != "" {
		md.Set(HeaderToken, token)
	}
	return md
}

func firstMD(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
