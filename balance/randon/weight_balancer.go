package randon

import (
	"math/rand"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type WeightBalancer struct {
	connList    []*weightConn
	count       int
	totalWeight int32
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.count == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	num := rand.Intn(int(b.totalWeight))
	for _, c := range b.connList {
		num -= int(c.weight)
		if num <= 0 {
			return balancer.PickResult{
				SubConn: c.c,
			}, nil
		}
	}
	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
}

type WeightBuilder struct {
}

func (b WeightBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connList := make([]*weightConn, len(info.ReadySCs))
	var totalWeight int32
	for sunConn, subInfo := range info.ReadySCs {
		weight, ok := subInfo.Address.Attributes.Value("weight").(int32)
		if !ok {
			panic("weight is invalid")
		}
		totalWeight += weight
		connList = append(connList, &weightConn{
			c:      sunConn,
			weight: weight,
		})
	}
	return &WeightBalancer{
		connList:    connList,
		count:       len(connList),
		totalWeight: totalWeight,
	}
}

type weightConn struct {
	c           balancer.SubConn
	weight      int32
}
