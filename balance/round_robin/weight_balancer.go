package roundrobin

import (
	"math"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type WeightBalancer struct {
	connList []*weightConn
	mu       sync.Mutex
	count    int32
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.count == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight int32
	var pick *weightConn
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, c := range b.connList {
		totalWeight += c.efficientWeight                         // 1
		c.currentWeight += c.efficientWeight                     // 2
		if pick == nil || c.currentWeight > pick.currentWeight { // pick the max currentWeight
			pick = c
		}
	}
	if pick == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	pick.currentWeight -= totalWeight // 3
	return balancer.PickResult{
		SubConn: pick.c,
		Done: func(info balancer.DoneInfo) {
			for {
				efficientWeight := atomic.LoadInt32(&pick.efficientWeight)
				if info.Err == nil && efficientWeight == math.MaxInt32 {
					return
				}
				if info.Err != nil && efficientWeight == 0 {
					return
				}
				newEfficientWeight := efficientWeight
				if info.Err != nil {
					newEfficientWeight--
				} else {
					newEfficientWeight++
				}
				if atomic.CompareAndSwapInt32(&pick.efficientWeight, efficientWeight, newEfficientWeight) {
					return
				}
			}
			// 方案2 直接加锁 优化：子锁 以 conn 为单位加锁 而不是 balancer
		},
	}, nil
}

type WeightBuilder struct {
}

func (b WeightBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connList := make([]*weightConn, len(info.ReadySCs))
	for sunConn, subInfo := range info.ReadySCs {
		weight, ok := subInfo.Address.Attributes.Value("weight").(int32)
		if !ok {
			panic("weight is invalid")
		}
		connList = append(connList, &weightConn{
			c:               sunConn,
			weight:          weight,
			efficientWeight: weight,
			currentWeight:   weight,
		})
	}
	return &WeightBalancer{
		connList: connList,
		count:    int32(len(connList)),
	}
}

type weightConn struct {
	c               balancer.SubConn
	efficientWeight int32
	currentWeight   int32
	weight          int32
}
