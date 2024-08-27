package go_micro

import (
	"context"
	"github.com/LXJ0000/go-micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	name            string
	registry        registry.Registry
	registerTimeout time.Duration
	weight int32
	*grpc.Server
}

func NewServer(name string, opt ...ServerOption) *Server {
	s := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registerTimeout: time.Second * 10,
	}
	for _, o := range opt {
		o(s)
	}
	return s
}

func (s *Server) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		if err = s.registry.Register(ctx, registry.ServiceInstance{
			Name: s.name,
			Addr: listener.Addr().String(),
		}); err != nil {
			return err
		}
		defer func() {
			_ = s.registry.Close()
		}()
	}
	return s.Serve(listener)
}

func (s *Server) Close() error {
	if s.registry != nil {
		if err := s.registry.Close(); err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}

func ServerWithRegister(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}

func ServerWithWeight(weight int) ServerOption {
	return func(server *Server) {
		server.weight = int32(weight)
	}
}