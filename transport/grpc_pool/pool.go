package grpc_pool

import (
	"errors"

	"google.golang.org/grpc"
)

// ErrClosed 连接池已经关闭Error
var ErrClosed = errors.New("pool is closed")

// Pool 基本方法
type Pool interface {
	Get() (*grpc.ClientConn, error)

	Put(*grpc.ClientConn) error

	Close(*grpc.ClientConn) error

	Release()

	Len() int
}
