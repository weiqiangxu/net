package grpc_pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ErrMaxActiveConnReached 连接池超限
var ErrMaxActiveConnReached = errors.New("MaxActiveConnReached")

// Config 连接池相关配置
type Config struct {
	// 连接池中拥有的最小连接数
	InitialCap int
	// 最大并发存活连接数
	MaxCap int
	// 最大空闲连接
	MaxIdle int
	// 生成连接的方法
	Factory func() (*grpc.ClientConn, error)
	// 关闭连接的方法
	Close func(*grpc.ClientConn) error
	// 检查连接是否有效的方法
	Ping func(*grpc.ClientConn) error
	// 连接最大空闲时间，超过该事件则将失效
	IdleTimeout time.Duration
}

type connReq struct {
	c *Conn
}

// channelPool 存放连接信息
type channelPool struct {
	mu                 sync.RWMutex
	connections        chan *Conn
	factory            func() (*grpc.ClientConn, error)
	close              func(*grpc.ClientConn) error
	ping               func(*grpc.ClientConn) error
	idleTimeout        time.Duration
	maxActive          int
	openingConnections int
	connReqs           []chan connReq
}

// Conn connection of grpc
type Conn struct {
	c *grpc.ClientConn
	t time.Time
}

// New 初始化连接
func New(poolConfig *Config) (Pool, error) {
	if !(poolConfig.InitialCap <= poolConfig.MaxIdle && poolConfig.MaxCap >= poolConfig.MaxIdle && poolConfig.InitialCap >= 0) {
		return nil, errors.New("invalid capacity settings")
	}
	if poolConfig.Factory == nil {
		return nil, errors.New("invalid factory func settings")
	}
	if poolConfig.Close == nil {
		return nil, errors.New("invalid close func settings")
	}

	c := &channelPool{
		connections:        make(chan *Conn, poolConfig.MaxIdle),
		factory:            poolConfig.Factory,
		close:              poolConfig.Close,
		idleTimeout:        poolConfig.IdleTimeout,
		maxActive:          poolConfig.MaxCap,
		openingConnections: poolConfig.InitialCap,
	}

	if poolConfig.Ping != nil {
		c.ping = poolConfig.Ping
	}

	for i := 0; i < poolConfig.InitialCap; i++ {
		connection, err := c.factory()
		if err != nil {
			c.Release()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.connections <- &Conn{c: connection, t: time.Now()}
	}

	return c, nil
}

// getConnection 获取所有连接
func (c *channelPool) getConnection() chan *Conn {
	c.mu.Lock()
	connections := c.connections
	c.mu.Unlock()
	return connections
}

// Get 从pool中取一个连接
func (c *channelPool) Get() (*grpc.ClientConn, error) {
	connections := c.getConnection()
	if connections == nil {
		return nil, ErrClosed
	}
	for {
		select {
		case wrapConn := <-connections:
			if wrapConn == nil {
				return nil, ErrClosed
			}
			// 判断是否超时，超时则丢弃
			if timeout := c.idleTimeout; timeout > 0 {
				if wrapConn.t.Add(timeout).Before(time.Now()) {
					// 丢弃并关闭该连接
					_ = c.Close(wrapConn.c)
					continue
				}
			}
			// 判断是否失效，失效则丢弃，如果用户没有设定 ping 方法，就不检查
			if c.ping != nil {
				if err := c.Ping(wrapConn.c); err != nil {
					_ = c.Close(wrapConn.c)
					continue
				}
			}
			return wrapConn.c, nil
		default:
			c.mu.Lock()
			if c.openingConnections >= c.maxActive {
				req := make(chan connReq, 1)
				c.connReqs = append(c.connReqs, req)
				c.mu.Unlock()
				ret, ok := <-req
				if !ok {
					return nil, ErrMaxActiveConnReached
				}
				if timeout := c.idleTimeout; timeout > 0 {
					if ret.c.t.Add(timeout).Before(time.Now()) {
						// 丢弃并关闭该连接
						_ = c.Close(ret.c.c)
						continue
					}
				}
				return ret.c.c, nil
			}
			if c.factory == nil {
				c.mu.Unlock()
				return nil, ErrClosed
			}
			conn, err := c.factory()
			if err != nil {
				c.mu.Unlock()
				return nil, err
			}
			c.openingConnections++
			c.mu.Unlock()
			return conn, nil
		}
	}
}

// Put 将连接放回pool中
func (c *channelPool) Put(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.Lock()

	if c.connections == nil {
		c.mu.Unlock()
		return c.Close(conn)
	}

	if l := len(c.connReqs); l > 0 {
		req := c.connReqs[0]
		copy(c.connReqs, c.connReqs[1:])
		c.connReqs = c.connReqs[:l-1]
		req <- connReq{
			c: &Conn{c: conn, t: time.Now()},
		}
		c.mu.Unlock()
		return nil
	} else {
		select {
		case c.connections <- &Conn{c: conn, t: time.Now()}:
			c.mu.Unlock()
			return nil
		default:
			c.mu.Unlock()
			// 连接池已满，直接关闭该连接
			return c.Close(conn)
		}
	}
}

// Close 关闭单条连接
func (c *channelPool) Close(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.close == nil {
		return nil
	}
	c.openingConnections--
	return c.close(conn)
}

// Ping 检查单条连接是否有效
func (c *channelPool) Ping(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	return c.ping(conn)
}

// Release 释放连接池中所有连接
func (c *channelPool) Release() {
	c.mu.Lock()
	connections := c.connections
	c.connections = nil
	c.factory = nil
	c.ping = nil
	c.close = nil
	c.mu.Unlock()

	if connections == nil {
		return
	}

	close(connections)
	for wrapConn := range connections {
		_ = c.close(wrapConn.c)
	}
}

// Len 连接池中已有的连接
func (c *channelPool) Len() int {
	return len(c.getConnection())
}
