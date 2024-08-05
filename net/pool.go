package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	freeConn     chan *conn // 空闲连接队列
	connRequests []request  // 等待连接的请求队列
	maxIdleCount int        // 最大空闲连接数
	maxOpen      int        // 最大连接数
	numOpen      int        // 当前打开的连接数 = 空闲连接数 + 非空闲连接数
	initOpen     int        // 初始打开的连接数

	maxIdleTime time.Duration // 连接最大空闲时间

	factory func() (net.Conn, error) // 连接工厂函数

	mu sync.RWMutex
}

type conn struct {
	net.Conn
	lastActiveAt time.Time
}

type request struct {
	ch chan net.Conn
}

func NewPool(initOpen, maxIdleCount, maxOpen int, maxIdleTime time.Duration, factory func() (net.Conn, error)) (*Pool, error) {
	if initOpen > maxIdleCount || maxIdleCount > maxOpen {
		return nil, errors.New("initOpen must be less than maxIdleCount and maxIdleCount must be less than maxOpen")
	}
	freeConn := make(chan *conn, maxIdleCount)
	for i := 0; i < initOpen; i++ {
		c, err := factory()
		if err != nil {
			return nil, err
		}
		freeConn <- &conn{c, time.Now()}
	}
	return &Pool{
		freeConn:     make(chan *conn, maxIdleCount),
		maxIdleCount: maxIdleCount,
		maxOpen:      maxOpen,
		maxIdleTime:  maxIdleTime,
		numOpen:      initOpen,
		factory:      factory,
	}, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done(): // 超时
		return nil, ctx.Err()
	default:
	}
	for { // 循环从空闲连接中获取连接
		select {
		case c := <-p.freeConn: // 从空闲连接中获取
			if c == nil {
				return nil, errors.New("connection pool is closed")
			}
			if p.maxIdleTime > 0 && time.Since(c.lastActiveAt) > p.maxIdleTime {
				_ = c.Conn.Close()
				continue
			}
			return c.Conn, nil
		default: // 空闲连接已用完，创建新的连接或者等待连接释放
			p.mu.Lock()
			if p.numOpen >= p.maxOpen { // 连接数已达到最大值 等待连接释放
				req := request{make(chan net.Conn, 1)}
				p.connRequests = append(p.connRequests, req)
				p.mu.Unlock() // 阻塞之前解锁
				select {
				case <-ctx.Done():
					// way1: 删除连接
					// way2: 转发连接
					go func() {
						c := <-req.ch
						_ = p.Put(context.Background(), c)
					}()
					return nil, ctx.Err()
				case c := <-req.ch: // 等待连接释放
					return c, nil // 连接释放后直接返回 该连接刚刚才被使用 无需检测
				}
			}
			// 连接数未达到最大值 创建新的连接
			c, err := p.factory()
			if err != nil {
				p.mu.Unlock()
				return nil, err
			}
			p.numOpen++
			p.mu.Unlock()
			return c, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, c net.Conn) error {
	p.mu.Lock()
	if len(p.connRequests) != 0 { // 等待队列不为空 即有阻塞的请求
		req := p.connRequests[0]
		p.connRequests = p.connRequests[1:]
		p.mu.Unlock() // 阻塞之前解锁
		req.ch <- c
		return nil
	}
	p.mu.Unlock()
	// 等待队列为空
	newConn := &conn{
		Conn: c, lastActiveAt: time.Now(),
	}
	select {
	case p.freeConn <- newConn: // 空闲连接未满，将连接放入空闲连接池
	default: // 空闲连接已满
		_ = c.Close()
		p.mu.Lock()
		p.numOpen--
		p.mu.Unlock()
	}
	return nil
}
