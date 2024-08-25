package go_micro

import (
	"context"
	"fmt"
	roundrobin "github.com/LXJ0000/go-micro/balance/round_robin"
	"github.com/LXJ0000/go-micro/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"time"
)

type ClientOption func(client *Client)

type Client struct {
	insecure bool
	registry registry.Registry
}

func NewClient(opt ...ClientOption) *Client {
	c := &Client{}
	for _, o := range opt {
		o(c)
	}
	return c
}

func WithClientInsecure() ClientOption {
	return func(client *Client) {
		client.insecure = true
	}
}

func WithClientRegistry(registry registry.Registry) ClientOption {
	return func(client *Client) {
		client.registry = registry
	}
}

func (c *Client) Dial(ctx context.Context, service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.registry != nil {
		builder := NewGRPCResolverBuilder(c.registry, time.Second*10)
		opts = append(opts, grpc.WithResolvers(builder))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	balancer.Register(base.NewBalancerBuilder("go_micro_round_robin", &roundrobin.Builder{}, base.Config{HealthCheck: true}))
	balance := grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "go_micro_round_robin"}`)
	opts = append(opts, balance)
	return grpc.DialContext(ctx, fmt.Sprintf("grpc:///%s", service), opts...)
}
