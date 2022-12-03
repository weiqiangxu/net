package http

import "github.com/gin-gonic/gin"

type ServerOption func(*Server)

func WithMiddleware(h ...gin.HandlerFunc) ServerOption {
	return func(server *Server) {
		server.handlersChain = h
	}
}

func WithAddress(addr string) ServerOption {
	return func(server *Server) {
		server.address = addr
	}
}

func WithPrometheus(enablePrometheus bool) ServerOption {
	return func(server *Server) {
		server.prometheus = enablePrometheus
	}
}

func WithProfile(profile bool) ServerOption {
	return func(server *Server) {
		server.profile = profile
	}
}

func WithTracing(tracing bool) ServerOption {
	return func(server *Server) {
		server.tracing = tracing
	}
}
