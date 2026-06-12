package rpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type noopManager struct{}

func (n *noopManager) Enabled() bool { return false }

func (n *noopManager) Server() *grpc.Server { return nil }

func (n *noopManager) RegisterServices(_ ...ServiceRegistrar) {}

func (n *noopManager) Client() *ClientManager { return nil }

func (n *noopManager) GrpcPort() int { return 0 }

func (n *noopManager) Listen() (net.Listener, error) { return nil, nil }

func (n *noopManager) Start(_ context.Context, _ net.Listener) error { return nil }

func (n *noopManager) GracefulStop() {}

func (n *noopManager) Shutdown(_ context.Context) error { return nil }
