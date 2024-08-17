package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LXJ0000/go-micro/registry"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
)

type Registry struct {
	c       *clientv3.Client
	session *concurrency.Session
	mu      sync.Mutex
	cancels []func()
}

func NewRegister(c *clientv3.Client) (*Registry, error) {
	session, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:       c,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, instance registry.ServiceInstance) error {
	value, err := json.Marshal(instance)
	if err != nil {
		return err
	}
	_, err = r.c.Put(ctx, r.instanceKey(instance), string(value), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, instance registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(instance))
	return err
}

func (r *Registry) ListService(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	resp, err := r.c.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	instances := make([]registry.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var instance registry.ServiceInstance
		if err = json.Unmarshal(kv.Value, &instance); err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mu.Unlock()
	ctx = clientv3.WithRequireLeader(ctx)
	event := r.c.Watch(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	ch := make(chan registry.Event)
	go func() {
		for {
			select {
			case e := <-event:
				if e.Err() != nil || e.Canceled {
					return
				}
				for range e.Events {
					ch <- registry.Event{}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	return r.session.Close()
}

func (r *Registry) instanceKey(s registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", s.Name, s.Addr)
}

func (r *Registry) serviceKey(serviceName string) string {
	return fmt.Sprintf("/micro/%s", serviceName)
}
