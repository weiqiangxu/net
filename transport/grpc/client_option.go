package grpc

import "google.golang.org/grpc"

// clientOptions is gRPC Client
type clientOptions struct {
	endpoint     string
	interceptors []grpc.UnaryClientInterceptor
	grpcOpts     []grpc.DialOption
	tracing      bool
	insecure     bool
}

// ClientOption is gRPC client option.
type ClientOption func(o *clientOptions)

// WithEndpoint with client endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for unary RPCs.
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.interceptors = in
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcOpts = opts
	}
}

// WithTracing trace
func WithTracing(opt bool) ClientOption {
	return func(o *clientOptions) {
		o.tracing = opt
	}
}
