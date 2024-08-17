package grpc

import (
	"context"
	go_micro "github.com/LXJ0000/go-micro"
	"github.com/LXJ0000/go-micro/proto/gen"
	"github.com/LXJ0000/go-micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewRegister(c)
	if err != nil {
		panic(err)
	}
	client := go_micro.NewClient(go_micro.WithClientRegistry(r), go_micro.WithClientInsecure())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cc, err := client.Dial(ctx, "user-service")
	if err != nil {
		t.Fatal(err)
	}
	userClient := gen.NewUserServiceClient(cc)

	resp, err := userClient.GetByID(ctx, &gen.GetByIDReq{Id: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
