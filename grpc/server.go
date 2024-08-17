package grpc

import (
	"context"
	"fmt"
	"github.com/LXJ0000/go-micro/proto/gen"
)

type UserServer struct {
	gen.UnimplementedUserServiceServer
}

func (s UserServer) GetByID(ctx context.Context, req *gen.GetByIDReq) (*gen.GetByIDResp, error) {
	fmt.Println(req)
	return &gen.GetByIDResp{
		User: &gen.User{
			Id:   1,
			Name: "lxj",
		},
	}, nil
}
