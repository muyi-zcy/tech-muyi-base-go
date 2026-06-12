package rpc

import (
	"context"
	"net"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"google.golang.org/grpc"
)

// ServiceRegistrar 业务服务 gRPC 注册回调
type ServiceRegistrar func(*grpc.Server)

// Manager gRPC 管理器（可插拔）
type Manager interface {
	Enabled() bool
	Server() *grpc.Server
	RegisterServices(registrars ...ServiceRegistrar)
	Client() *ClientManager
	GrpcPort() int
	Listen() (net.Listener, error)
	Start(ctx context.Context, lis net.Listener) error
	GracefulStop()
	Shutdown(ctx context.Context) error
}

var globalManager Manager = &noopManager{}

// Init 根据配置初始化 RPC Manager
func Init(cfg *config.Config, registry nacos.Registry) Manager {
	m := initManager(cfg, registry)
	globalManager = m
	return m
}

// InitFromConfig 从全局配置初始化
func InitFromConfig(registry nacos.Registry) Manager {
	return Init(config.GetConfig(), registry)
}

// GetManager 获取全局 Manager
func GetManager() Manager {
	return globalManager
}

// SetManager 设置 Manager（测试用）
func SetManager(m Manager) {
	globalManager = m
}

func initManager(cfg *config.Config, registry nacos.Registry) Manager {
	if cfg == nil || !cfg.Plugins.Rpc.Enabled {
		return &noopManager{}
	}
	rpcCfg := cfg.Plugins.Rpc
	if rpcCfg.Registry == "nacos" && !registry.Enabled() {
		myLogger.Warn("Nacos 未启用或初始化失败，RPC 自动降级为 static 模式")
		rpcCfg.Registry = "static"
	}
	return newGrpcManager(cfg, registry, rpcCfg)
}
