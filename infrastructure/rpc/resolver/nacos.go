package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"google.golang.org/grpc/resolver"
)

const nacosScheme = "nacos"

// NacosBuilder nacos:///serviceName
type NacosBuilder struct {
	Naming nacos.NamingClient
	Group  string
}

func NewNacosBuilder(naming nacos.NamingClient, group string) *NacosBuilder {
	return &NacosBuilder{Naming: naming, Group: group}
}

func (b *NacosBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	if b.Naming == nil {
		return nil, fmt.Errorf("nacos resolver: naming client is nil")
	}
	serviceName := target.URL.Host
	if serviceName == "" {
		serviceName = stringsTrimPrefix(target.Endpoint(), "/")
	}
	r := &nacosResolver{
		cc:          cc,
		naming:      b.Naming,
		group:       b.Group,
		serviceName: serviceName,
	}
	if err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

func (b *NacosBuilder) Scheme() string { return nacosScheme }

type nacosResolver struct {
	cc          resolver.ClientConn
	naming      nacos.NamingClient
	group       string
	serviceName string
	cancel      context.CancelFunc
	mu          sync.Mutex
}

func (r *nacosResolver) start() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	instances, err := r.naming.SelectInstances(r.serviceName, r.group)
	if err != nil {
		return err
	}
	r.updateAddresses(instances)

	return r.naming.Subscribe(r.serviceName, r.group, func(instances []nacos.ServiceInstance) {
		select {
		case <-ctx.Done():
			return
		default:
			r.updateAddresses(instances)
		}
	})
}

func (r *nacosResolver) updateAddresses(instances []nacos.ServiceInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	addrs := make([]resolver.Address, 0, len(instances))
	for _, inst := range instances {
		if !inst.Healthy {
			continue
		}
		addrs = append(addrs, resolver.Address{
			Addr:     fmt.Sprintf("%s:%d", inst.IP, inst.Port),
			Metadata: inst.Metadata,
		})
	}
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *nacosResolver) ResolveNow(resolver.ResolveNowOptions) {
	instances, err := r.naming.SelectInstances(r.serviceName, r.group)
	if err == nil {
		r.updateAddresses(instances)
	}
}

func (r *nacosResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	_ = r.naming.Unsubscribe(r.serviceName, r.group)
}

// RegisterNacos 注册 nacos resolver
func RegisterNacos(naming nacos.NamingClient, group string) {
	resolver.Register(NewNacosBuilder(naming, group))
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
