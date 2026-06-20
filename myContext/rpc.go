package myContext

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const SourceService = "sourceService"

// WithMetadata 从 gRPC metadata 恢复上下文；仅信任 traceId 与 token，ssoId 由鉴权层写入。
func WithMetadata(ctx context.Context, md metadata.MD) context.Context {
	if v := firstMD(md, HeaderTraceId); v != "" {
		ctx = context.WithValue(ctx, keyTrace, v)
	}
	if v := firstMD(md, HeaderToken); v != "" {
		ctx = context.WithValue(ctx, keyToken, v)
	}
	if v := firstMD(md, HeaderSourceService); v != "" {
		ctx = context.WithValue(ctx, keySourceService, v)
	}
	return ctx
}

// MetadataFromContext 从 context 提取 metadata；ssoId 仅在鉴权后随 token 一并传播。
func MetadataFromContext(ctx context.Context) metadata.MD {
	md := metadata.MD{}
	if traceId := TryGetTraceId(ctx); traceId != "" {
		md.Set(HeaderTraceId, traceId)
	}
	if token := TryGetToken(ctx); token != "" {
		md.Set(HeaderToken, token)
	}
	if ssoId := TryGetSsoId(ctx); ssoId != "" {
		md.Set(HeaderSsoId, ssoId)
	}
	return md
}

func firstMD(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
