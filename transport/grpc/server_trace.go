package grpc

import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

// TraceDecorator decorator a trace func to server
func (s *Server) TraceDecorator() {
	if s.tracing {
		s.unaryInterceptor = append(s.unaryInterceptor, otelgrpc.UnaryServerInterceptor())
		s.streamInterceptor = append(s.streamInterceptor, otelgrpc.StreamServerInterceptor())
	}
}
