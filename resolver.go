package go_micro

import (
	"context"
	"github.com/LXJ0000/go-micro/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

type GRPCResolverBuilder struct {
	registry       registry.Registry
	contextTimeout time.Duration
}

func NewGRPCResolverBuilder(registry registry.Registry, contextTimeout time.Duration) *GRPCResolverBuilder {
	return &GRPCResolverBuilder{registry: registry, contextTimeout: contextTimeout}
}

func (g *GRPCResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &GRPCResolver{
		target:         target,
		cc:             cc,
		registry:       g.registry,
		contextTimeout: g.contextTimeout,
	}
	r.resolve()
	go r.watch()
	return r, nil
}

func (g *GRPCResolverBuilder) Scheme() string {
	return "grpc"
}

type GRPCResolver struct {
	target         resolver.Target
	cc             resolver.ClientConn
	registry       registry.Registry
	contextTimeout time.Duration
	close          chan struct{}
}

func (g *GRPCResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *GRPCResolver) resolve(options ...resolver.ResolveNowOptions) {
	ctx, cancel := context.WithTimeout(context.Background(), g.contextTimeout)
	defer cancel()
	instances, err := g.registry.ListService(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	addresses := make([]resolver.Address, 0, len(instances))
	for _, i := range instances {
		addresses = append(addresses, resolver.Address{Addr: i.Addr})
	}
	if err := g.cc.UpdateState(resolver.State{Addresses: addresses}); err != nil {
		g.cc.ReportError(err)
	}
}

func (g *GRPCResolver) watch(options ...resolver.ResolveNowOptions) {
	events, err := g.registry.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			g.resolve()
		case <-g.close:
			return
		}
	}
}

func (g *GRPCResolver) Close() {
	close(g.close)
}
