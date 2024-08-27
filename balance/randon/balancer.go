package randon

import (
	"math/rand"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Balancer struct {
	connList []balancer.SubConn
	count    int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.count == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	num := rand.Intn(len(b.connList))

	return balancer.PickResult{
		SubConn: b.connList[num],
	}, nil
}

type Builder struct {
}

func (b Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connList := make([]balancer.SubConn, len(info.ReadySCs))
	for sunConn := range info.ReadySCs {
		connList = append(connList, sunConn)
	}
	return &Balancer{
		connList: connList,
		count:    len(connList),
	}
}
