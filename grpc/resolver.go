package grpc

import "google.golang.org/grpc/resolver"

type Builder struct {
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &Resolver{
		target: target,
		cc:     cc,
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (b *Builder) Scheme() string {
	return "grpc"
}

type Resolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	if err := r.cc.UpdateState(
		resolver.State{
			Addresses: []resolver.Address{
				{Addr: "localhost:8080"},
			}},
	); err != nil {
		r.cc.ReportError(err)
	}
}

func (r *Resolver) Close() {
	//TODO implement me
	panic("implement me")
}
