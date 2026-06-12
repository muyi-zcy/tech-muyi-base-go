package interceptor

import (
	"context"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ContextInject Client：context → outgoing metadata
func ContextInject(sourceService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		pairs := make([]string, 0, 8)
		if traceId := myContext.TryGetTraceId(ctx); traceId != "" {
			pairs = append(pairs, myContext.HeaderTraceId, traceId)
		}
		if ssoId := myContext.TryGetSsoId(ctx); ssoId != "" {
			pairs = append(pairs, myContext.HeaderSsoId, ssoId)
		}
		if token := myContext.TryGetToken(ctx); token != "" {
			pairs = append(pairs, myContext.HeaderToken, token)
		}
		if sourceService != "" {
			pairs = append(pairs, myContext.HeaderSourceService, sourceService)
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ContextExtract Server：incoming metadata → context
func ContextExtract() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = myContext.WithMetadata(ctx, md)
		}
		return handler(ctx, req)
	}
}

// Logging Server 侧请求日志
func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		myLogger.Info("gRPC request",
			zap.String("method", info.FullMethod),
			zap.String("traceId", myContext.TryGetTraceId(ctx)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}

// ClientLogging Client 侧请求日志
func ClientLogging() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		myLogger.Info("gRPC client",
			zap.String("method", method),
			zap.String("traceId", myContext.TryGetTraceId(ctx)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return err
	}
}
