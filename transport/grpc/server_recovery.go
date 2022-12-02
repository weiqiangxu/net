package grpc

import (
	"context"
	"fmt"

	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/weiqiangxu/common-config/logger"
	"google.golang.org/grpc/metadata"
)

// RecoveryDecorator decorator a recovery func to server
func (s *Server) RecoveryDecorator() {
	if s.recovery {
		grpcRecoveryOpts := []grpcRecovery.Option{
			grpcRecovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
				md, ok := metadata.FromIncomingContext(ctx)
				if ok {
					logger.Errorf("gRPC metadata: %v panic info: %v", md, p)
				} else {
					logger.Error(p)
				}
				return fmt.Errorf("context panic triggered: %v", p)
			}),
		}
		s.unaryInterceptor = append(s.unaryInterceptor, grpcRecovery.UnaryServerInterceptor(grpcRecoveryOpts...))
		s.streamInterceptor = append(s.streamInterceptor, grpcRecovery.StreamServerInterceptor(grpcRecoveryOpts...))
	}
}
