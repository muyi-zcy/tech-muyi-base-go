package nacos

import (
	"context"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
)

// Registry Nacos 注册与发现接口（可插拔）
type Registry interface {
	Enabled() bool
	Register(ctx context.Context, grpcPort int, metadata map[string]string) error
	Deregister(ctx context.Context) error
	NamingClient() NamingClient
}

// NamingClient 供 gRPC Resolver 使用的命名服务抽象
type NamingClient interface {
	Subscribe(serviceName, groupName string, callback func(instances []ServiceInstance)) error
	Unsubscribe(serviceName, groupName string) error
	SelectInstances(serviceName, groupName string) ([]ServiceInstance, error)
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	IP       string
	Port     uint64
	Weight   float64
	Healthy  bool
	Metadata map[string]string
}

var globalRegistry Registry = &noopRegistry{}

// Init 根据配置初始化 Nacos（enabled=false 时使用 noop，不阻断启动）
func Init(cfg *config.Config) Registry {
	reg, err := InitWithError(cfg)
	if err != nil {
		myLogger.Warn("Nacos 初始化失败，降级为 noop", zap.Error(err))
	}
	return reg
}

// InitFromConfig 从全局配置初始化
func InitFromConfig() Registry {
	return Init(config.GetConfig())
}

// GetRegistry 获取全局 Registry
func GetRegistry() Registry {
	return globalRegistry
}

// SetRegistry 设置 Registry（测试用）
func SetRegistry(r Registry) {
	globalRegistry = r
}
