package grpc

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
)

func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{}
	for _, o := range opts {
		o(&options)
	}
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor
	if len(options.interceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, options.interceptors...)
	}
	if options.tracing {
		unaryInterceptors = append(unaryInterceptors, otelgrpc.UnaryClientInterceptor())
		streamInterceptors = append(streamInterceptors, otelgrpc.StreamClientInterceptor())
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": %q}`, roundrobin.Name)),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}
	if options.insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}
