package rpc

import (
	"sync"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc/interceptor"
)

// ClientManager gRPC 客户端连接池
type ClientManager struct {
	cfg            config.RpcClientConfig
	registry       string
	staticAddrs    map[string]string
	sourceService  string
	defaultTimeout time.Duration
	mu             sync.RWMutex
	conns          map[string]*grpc.ClientConn
}

func newClientManager(cfg *config.Config, rpcCfg config.RpcConfig) *ClientManager {
	timeout := time.Duration(rpcCfg.Client.DefaultTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	services := rpcCfg.Client.Services
	if services == nil {
		services = map[string]string{}
	}
	sourceService := cfg.Plugins.Nacos.ServiceName
	if sourceService == "" {
		sourceService = cfg.AppName
	}
	return &ClientManager{
		cfg:            rpcCfg.Client,
		registry:       rpcCfg.Registry,
		staticAddrs:    rpcCfg.Static,
		sourceService:  sourceService,
		defaultTimeout: timeout,
		conns:          make(map[string]*grpc.ClientConn),
	}
}

// GetConn 获取到目标服务的 gRPC 连接（serviceKey 为配置键，如 xi.user）
func (m *ClientManager) GetConn(serviceKey string) (*grpc.ClientConn, error) {
	m.mu.RLock()
	if conn, ok := m.conns[serviceKey]; ok {
		m.mu.RUnlock()
		return conn, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.conns[serviceKey]; ok {
		return conn, nil
	}

	target, err := m.buildTarget(serviceKey)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptor.ContextInject(m.sourceService),
			interceptor.ClientLogging(),
			interceptor.ClientErrorDecode(),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "连接 gRPC 服务 %s 失败", serviceKey)
	}
	m.conns[serviceKey] = conn
	myLogger.Info("gRPC 客户端已连接", zap.String("serviceKey", serviceKey), zap.String("target", target))
	return conn, nil
}

// DefaultTimeout 默认 RPC 超时
func (m *ClientManager) DefaultTimeout() time.Duration {
	return m.defaultTimeout
}

// Close 关闭所有连接
func (m *ClientManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var lastErr error
	for key, conn := range m.conns {
		if err := conn.Close(); err != nil {
			lastErr = err
			myLogger.Warn("关闭 gRPC 连接失败", zap.String("serviceKey", key), zap.Error(err))
		}
	}
	m.conns = make(map[string]*grpc.ClientConn)
	return lastErr
}

func (m *ClientManager) buildTarget(serviceKey string) (string, error) {
	serviceName := serviceKey
	if mapped, ok := m.cfg.Services[serviceKey]; ok && mapped != "" {
		serviceName = mapped
	}

	switch m.registry {
	case "static":
		addr := serviceName
		if mapped, ok := m.staticAddrs[serviceKey]; ok && mapped != "" {
			addr = mapped
		} else if mapped, ok := m.staticAddrs[serviceName]; ok && mapped != "" {
			addr = mapped
		}
		if addr == "" {
			return "", errors.Errorf("static 模式未配置服务 %s 的地址", serviceKey)
		}
		return "static:///" + addr, nil
	default:
		return "nacos:///" + serviceName, nil
	}
}
