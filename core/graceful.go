package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc"
	"github.com/muyi-zcy/tech-muyi-base-go/middleware"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RunOptions 启动选项
type RunOptions struct {
	// RegisterGrpc 业务服务注册 gRPC 实现（plugins.rpc.enabled=true 时生效）
	RegisterGrpc []rpc.ServiceRegistrar
}

// Run 启动 HTTP 服务（向后兼容，RPC/Nacos 插件按配置自动启用）
func (s *Starter) Run() error {
	return s.RunWithOptions(RunOptions{})
}

// RunWithOptions 启动 HTTP + gRPC（可插拔）并优雅退出
func (s *Starter) RunWithOptions(opts RunOptions) error {
	healthController := NewHealthCheckController()
	RegisterHealthCheckRoutes(s.Engine, healthController)
	s.Engine.NoRoute(middleware.NotFoundHandler())
	s.Engine.NoMethod(middleware.MethodNotAllowedHandler())

	httpPort := 8080
	if s.App.Config != nil && s.App.Config.Server.Port > 0 {
		httpPort = s.App.Config.Server.Port
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: s.Engine,
	}

	ctx := context.Background()
	rpcMgr := rpc.GetManager()

	if rpcMgr.Enabled() {
		if len(opts.RegisterGrpc) > 0 {
			rpcMgr.RegisterServices(opts.RegisterGrpc...)
		}
		lis, err := rpcMgr.Listen()
		if err != nil {
			return err
		}
		if err := rpcMgr.Start(ctx, lis); err != nil {
			return err
		}
		if reg := nacos.GetRegistry(); reg.Enabled() {
			if err := reg.Register(ctx, rpcMgr.GrpcPort(), nil); err != nil {
				myLogger.Warn("Nacos 注册失败，服务继续运行", zap.Error(err))
			}
		}
	}

	go func() {
		myLogger.Info("HTTP Server 启动",
			zap.String("name", s.App.Name),
			zap.Int("port", httpPort),
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			myLogger.Error("HTTP Server 异常退出", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	myLogger.Info("收到退出信号，开始优雅关闭...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if reg := nacos.GetRegistry(); reg.Enabled() {
		if err := reg.Deregister(shutdownCtx); err != nil {
			myLogger.Warn("Nacos 注销失败", zap.Error(err))
		}
	}
	if rpcMgr.Enabled() {
		rpcMgr.GracefulStop()
		if err := rpcMgr.Shutdown(shutdownCtx); err != nil {
			myLogger.Warn("RPC 关闭失败", zap.Error(err))
		}
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		myLogger.Warn("HTTP 关闭失败", zap.Error(err))
	}
	if err := s.App.Shutdown(); err != nil {
		return err
	}
	myLogger.Info("服务已优雅退出")
	return nil
}
