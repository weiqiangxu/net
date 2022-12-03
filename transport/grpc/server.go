package grpc

import (
	"context"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/weiqiangxu/net/transport"

	"github.com/weiqiangxu/net/tool"

	"github.com/weiqiangxu/common-config/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var _ transport.Server = (*Server)(nil)

const (
	DefaultNetProtocol = "tcp"
	DefaultNetAddress  = ":0"
	HealthcheckService = "grpc.health.v1.Health"
	SchemeOfGrpc       = "grpc"
)

type Server struct {
	*grpc.Server
	ctx               context.Context
	listener          net.Listener
	once              sync.Once
	err               error
	network           string
	address           string
	endpoint          *url.URL
	timeout           time.Duration
	unaryInterceptor  []grpc.UnaryServerInterceptor
	streamInterceptor []grpc.StreamServerInterceptor
	grpcOpts          []grpc.ServerOption
	health            *health.Server
	tracing           bool
	recovery          bool
}

func NewServer(opts ...ServerOption) *Server {
	server := &Server{network: DefaultNetProtocol, address: DefaultNetAddress, timeout: 1 * time.Second, health: health.NewServer()}
	for _, o := range opts {
		o(server)
	}
	server.TraceDecorator()
	server.RecoveryDecorator()
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(server.unaryInterceptor...),
		grpc.ChainStreamInterceptor(server.streamInterceptor...),
	}
	server.grpcOpts = append(server.grpcOpts, grpcOpts...)
	server.Server = grpc.NewServer(server.grpcOpts...)
	server.health.SetServingStatus(HealthcheckService, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server.Server, server.health)
	reflection.Register(server.Server)
	return server
}

func (s *Server) Start(ctx context.Context) error {
	if _, err := s.endpointListen(); err != nil {
		return err
	}
	s.ctx = ctx
	logger.Infof("[gRPC] server listening on: %s", s.listener.Addr().String())
	s.health.Resume()
	return s.Serve(s.listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.GracefulStop()
	s.health.Shutdown()
	logger.Info("[gRPC] server stopping")
	return nil
}

// endpointListen return a real address to registry endpoint
func (s *Server) endpointListen() (*url.URL, error) {
	s.once.Do(func() {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return
		}
		addr, err := tool.Extract(s.address, s.listener)
		if err != nil {
			err := lis.Close()
			if err != nil {
				logger.Errorf("close %s listener catch err=%v", s.address, err)
			}
			s.err = err
			return
		}
		s.listener = lis
		s.endpoint = &url.URL{Scheme: SchemeOfGrpc, Host: addr}
	})
	if s.err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}
