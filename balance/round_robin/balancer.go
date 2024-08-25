package roundrobin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	connList []balancer.SubConn
	index    int32
	count    int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.count == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	index := atomic.AddInt32(&b.index, 1)
	conn := b.connList[index%b.count] // 只读无需保护
	return balancer.PickResult{
		SubConn: conn,
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
		index:    -1,
		count:    int32(len(connList)),
	}
}
