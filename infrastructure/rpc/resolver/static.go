package resolver

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

const staticScheme = "static"

// StaticBuilder static:///host:port 或 static:///serviceKey（从 map 查地址）
type StaticBuilder struct {
	Addresses map[string]string
}

func NewStaticBuilder(addresses map[string]string) *StaticBuilder {
	return &StaticBuilder{Addresses: addresses}
}

func (b *StaticBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	addr := target.URL.Host
	if addr == "" {
		addr = strings.TrimPrefix(target.URL.Path, "/")
	}
	if addr == "" {
		addr = strings.TrimPrefix(target.Endpoint(), "/")
	}
	if mapped, ok := b.Addresses[addr]; ok {
		addr = mapped
	}
	if !strings.Contains(addr, ":") {
		return nil, fmt.Errorf("static resolver: invalid address %q", addr)
	}
	r := &staticResolver{
		cc:   cc,
		addr: addr,
	}
	r.start()
	return r, nil
}

func (b *StaticBuilder) Scheme() string { return staticScheme }

type staticResolver struct {
	cc   resolver.ClientConn
	addr string
}

func (r *staticResolver) start() {
	_ = r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{{Addr: r.addr}},
	})
}

func (r *staticResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *staticResolver) Close() {}

// RegisterStatic 注册 static resolver
func RegisterStatic(addresses map[string]string) {
	resolver.Register(NewStaticBuilder(addresses))
}
