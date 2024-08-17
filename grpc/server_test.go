package grpc

import (
	"github.com/LXJ0000/go-micro/proto/gen"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	userServer := &UserServer{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, userServer)
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	if err = server.Serve(lis); err != nil {
		t.Fatal(err)
	}
}
