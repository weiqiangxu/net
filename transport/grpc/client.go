package grpc

import (
	"context"
	"fmt"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/weiqiangxu/user/config"
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
	if options.prometheus {
		grpcPrometheus.EnableClientHandlingTimeHistogram(
			WithGrpcHistogramName(config.Conf.Application.Name, "grpc_seconds"),
		)
		list := []grpc.DialOption{
			grpc.WithUnaryInterceptor(grpcPrometheus.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpcPrometheus.StreamClientInterceptor),
		}
		grpcOpts = append(grpcOpts, list...)
	}
	if options.insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}
	if options.tracerInterceptor {
		list := []grpc.DialOption{
			grpc.WithUnaryInterceptor(ClientInterceptor(options.tracer)),
		}
		grpcOpts = append(grpcOpts, list...)
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}

// WithGrpcHistogramName change prometheus histogramName
func WithGrpcHistogramName(namespace string, name string) grpcPrometheus.HistogramOption {
	return func(o *prom.HistogramOpts) {
		o.Namespace = namespace
		o.Name = name
	}
}

// ClientInterceptor grpc client
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		parentSpan := ctx.Value("span")
		if parentSpan != nil {
			parentSpanContext := parentSpan.(opentracing.SpanContext)
			child := tracer.StartSpan(method, opentracing.ChildOf(parentSpanContext))
			defer child.Finish()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
