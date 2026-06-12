package nacos

import (
	"context"
)

type noopRegistry struct{}

func (n *noopRegistry) Enabled() bool { return false }

func (n *noopRegistry) Register(_ context.Context, _ int, _ map[string]string) error {
	return nil
}

func (n *noopRegistry) Deregister(_ context.Context) error {
	return nil
}

func (n *noopRegistry) NamingClient() NamingClient {
	return nil
}
