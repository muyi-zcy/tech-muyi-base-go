package core

import (
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc"
)

// RegisterGrpcServices 注册 gRPC 服务（需在 Run 之前调用）
func (s *Starter) RegisterGrpcServices(registrars ...rpc.ServiceRegistrar) {
	rpc.GetManager().RegisterServices(registrars...)
}

// GetRpcManager 获取 RPC Manager
func (s *Starter) GetRpcManager() rpc.Manager {
	return rpc.GetManager()
}

// GetNacosRegistry 获取 Nacos Registry
func (s *Starter) GetNacosRegistry() nacos.Registry {
	return nacos.GetRegistry()
}

// RunWithGrpc 启动并注册 gRPC 服务的便捷方法
func (s *Starter) RunWithGrpc(registrars ...rpc.ServiceRegistrar) error {
	return s.RunWithOptions(RunOptions{RegisterGrpc: registrars})
}
