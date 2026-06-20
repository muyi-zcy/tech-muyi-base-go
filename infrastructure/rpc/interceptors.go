package rpc

import (
	"sync"

	"google.golang.org/grpc"
)

var (
	extraUnaryServerInterceptors []grpc.UnaryServerInterceptor
	extraInterceptorMu           sync.RWMutex
)

// RegisterUnaryServerInterceptor 注册额外 Unary 拦截器。
// 须在 rpc.Init 之后、Manager.Start / Run 之前调用；插入在 ContextExtract 之后、Logging 之前。
func RegisterUnaryServerInterceptor(i grpc.UnaryServerInterceptor) {
	if i == nil {
		return
	}
	extraInterceptorMu.Lock()
	defer extraInterceptorMu.Unlock()
	extraUnaryServerInterceptors = append(extraUnaryServerInterceptors, i)
}

func appendExtraUnaryServerInterceptors(chain []grpc.UnaryServerInterceptor) []grpc.UnaryServerInterceptor {
	extraInterceptorMu.RLock()
	defer extraInterceptorMu.RUnlock()
	if len(extraUnaryServerInterceptors) == 0 {
		return chain
	}
	out := make([]grpc.UnaryServerInterceptor, 0, len(chain)+len(extraUnaryServerInterceptors))
	out = append(out, chain...)
	out = append(out, extraUnaryServerInterceptors...)
	return out
}
