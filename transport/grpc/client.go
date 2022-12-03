package grpc

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
)

// Dial rpc client dial an address
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.tracing {
		options.unaryInterceptors = append(options.unaryInterceptors, otelgrpc.UnaryClientInterceptor())
		options.streamInterceptors = append(options.streamInterceptors, otelgrpc.StreamClientInterceptor())
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": %q}`, roundrobin.Name)),
		grpc.WithChainUnaryInterceptor(options.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(options.streamInterceptors...),
	}
	if options.insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}
