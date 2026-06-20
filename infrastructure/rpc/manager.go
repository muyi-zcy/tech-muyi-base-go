package rpc

import (
	"context"
	"net"
	"strconv"
	"sync"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc/interceptor"
	rpcresolver "github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc/resolver"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type grpcManager struct {
	cfg         config.RpcConfig
	grpcPort    int
	maxRecvSize int
	server      *grpc.Server
	client      *ClientManager
	registrars  []ServiceRegistrar
	healthSrv   *health.Server
	listener    net.Listener
	buildOnce   sync.Once
	startOnce   sync.Once
}

var resolversRegistered sync.Once

func newGrpcManager(appCfg *config.Config, registry nacos.Registry, rpcCfg config.RpcConfig) *grpcManager {
	registerResolversOnce(registry, rpcCfg)

	if rpcCfg.Server.Port <= 0 {
		rpcCfg.Server.Port = 9080
	}
	maxRecv := rpcCfg.Server.MaxRecvMsgSize
	if maxRecv <= 0 {
		maxRecv = 4 << 20
	}

	mgr := &grpcManager{
		cfg:         rpcCfg,
		grpcPort:    rpcCfg.Server.Port,
		maxRecvSize: maxRecv,
		client:      newClientManager(appCfg, rpcCfg),
	}

	myLogger.Info("RPC 插件已启用",
		zap.Int("grpcPort", rpcCfg.Server.Port),
		zap.String("registry", rpcCfg.Registry),
	)
	return mgr
}

func (m *grpcManager) buildServer() {
	m.buildOnce.Do(func() {
		chain := []grpc.UnaryServerInterceptor{
			interceptor.Recovery(),
			interceptor.ContextExtract(),
		}
		chain = appendExtraUnaryServerInterceptors(chain)
		chain = append(chain,
			interceptor.Logging(),
			interceptor.ErrorMapping(),
		)

		opts := []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(chain...),
			grpc.MaxRecvMsgSize(m.maxRecvSize),
		}

		srv := grpc.NewServer(opts...)
		healthSrv := health.NewServer()
		healthpb.RegisterHealthServer(srv, healthSrv)
		healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

		if m.cfg.Server.EnableReflection {
			reflection.Register(srv)
		}

		m.server = srv
		m.healthSrv = healthSrv
	})
}

func registerResolversOnce(registry nacos.Registry, rpcCfg config.RpcConfig) {
	resolversRegistered.Do(func() {
		rpcresolver.RegisterStatic(rpcCfg.Static)
		if rpcCfg.Registry != "static" {
			if naming := registry.NamingClient(); naming != nil {
				group := config.GetNacosConfig().Group
				if group == "" {
					group = "XI_PLATFORM"
				}
				rpcresolver.RegisterNacos(naming, group)
			}
		}
	})
}

func (m *grpcManager) Enabled() bool { return true }

func (m *grpcManager) Server() *grpc.Server {
	m.buildServer()
	return m.server
}

func (m *grpcManager) RegisterServices(registrars ...ServiceRegistrar) {
	m.registrars = append(m.registrars, registrars...)
}

func (m *grpcManager) Client() *ClientManager { return m.client }

func (m *grpcManager) GrpcPort() int { return m.grpcPort }

func (m *grpcManager) Start(_ context.Context, lis net.Listener) error {
	var startErr error
	m.startOnce.Do(func() {
		m.buildServer()
		for _, reg := range m.registrars {
			reg(m.server)
		}
		m.listener = lis
		go func() {
			myLogger.Info("gRPC Server 启动", zap.String("addr", lis.Addr().String()))
			if err := m.server.Serve(lis); err != nil {
				myLogger.Error("gRPC Server 异常退出", zap.Error(err))
			}
		}()
	})
	return startErr
}

func (m *grpcManager) GracefulStop() {
	if m.server != nil {
		m.server.GracefulStop()
	}
}

func (m *grpcManager) Shutdown(_ context.Context) error {
	m.GracefulStop()
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// Listen 创建 gRPC 监听器
func (m *grpcManager) Listen() (net.Listener, error) {
	addr := net.JoinHostPort("", strconv.Itoa(m.grpcPort))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "gRPC 监听失败")
	}
	return lis, nil
}
