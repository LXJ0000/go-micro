package go_micro

import (
	_grpc "github.com/LXJ0000/go-micro/grpc"
	"github.com/LXJ0000/go-micro/proto/gen"
	"github.com/LXJ0000/go-micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := etcd.NewRegister(etcdClient)
	if err != nil {
		t.Fatal(err)
	}
	userServer := &_grpc.UserServer{}
	server := NewServer("user-service", ServerWithRegister(r))
	gen.RegisterUserServiceServer(server, userServer)
	if err = server.Run(":8080"); err != nil {
		t.Fatal(err)
	}
}
