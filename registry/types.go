package registry

import (
	"context"
	"io"
)

type Registry interface {
	// Register registers a new service
	Register(ctx context.Context, instance ServiceInstance) error
	// UnRegister unregisters a service
	UnRegister(ctx context.Context, instance ServiceInstance) error
	// ListService lists all services
	ListService(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	// Subscribe subscribes to a service
	Subscribe(serviceName string) (<-chan Event, error)
	io.Closer
}

type ServiceInstance struct {
	Name string
	Addr string
}

type Event struct {
	Type string
}
