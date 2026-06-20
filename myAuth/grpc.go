package myAuth

import (
	"context"

	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"google.golang.org/grpc"
)

type grpcAuthOptions struct {
	providerName  string
	enricherNames []string
}

// GRPCOption gRPC 鉴权拦截器配置。
type GRPCOption func(*grpcAuthOptions)

func WithGRPCProvider(name string) GRPCOption {
	return func(o *grpcAuthOptions) {
		o.providerName = name
	}
}

func WithGRPCEnrichers(names ...string) GRPCOption {
	return func(o *grpcAuthOptions) {
		o.enricherNames = append(o.enricherNames, names...)
	}
}

// GRPCAuthInterceptor 从 metadata token 校验并写入 Session/ssoId；不信任裸 x-sso-id。
func GRPCAuthInterceptor(opts ...GRPCOption) grpc.UnaryServerInterceptor {
	options := grpcAuthOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token := myContext.TryGetToken(ctx)
		if token == "" {
			return handler(ctx, req)
		}

		ensureInitialized()
		m := globalManager.(*manager)
		sess, err := m.LoadFromToken(ctx, token, options.providerName)
		if err != nil || sess == nil {
			return handler(ctx, req)
		}

		if err := runEnrichers(ctx, sess, options.enricherNames); err != nil {
			return nil, err
		}

		ctx = bindAuthContextGRPC(ctx, sess)
		return handler(ctx, req)
	}
}

// RegisterGRPCAuth 将 GRPCAuthInterceptor 注册到全局 gRPC 链。
// 须在 myAuth.Init / MustInitFromViper 之后、starter.Run 之前调用。
func RegisterGRPCAuth(opts ...GRPCOption) {
	rpc.RegisterUnaryServerInterceptor(GRPCAuthInterceptor(opts...))
}
