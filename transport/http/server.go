package http

import (
	"context"
	"net/http"
	"time"

	"github.com/weiqiangxu/net/transport"
	"github.com/weiqiangxu/user/config"

	"github.com/weiqiangxu/common-config/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"

	ginPprof "github.com/gin-contrib/pprof"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	DefaultHttpNetwork = "tcp"
	DefaultHttpAddress = ":0"
)

var _ transport.Server = (*Server)(nil)

type Server struct {
	gin           *gin.Engine
	httpServer    *http.Server
	address       string
	network       string
	handlersChain []gin.HandlerFunc
	prometheus    bool
	profile       bool
	tracing       bool
}

func NewServer(opts ...ServerOption) *Server {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	g := gin.New()
	srv := &Server{
		network: DefaultHttpNetwork,
		address: DefaultHttpAddress,
	}
	for _, o := range opts {
		o(srv)
	}
	if srv.prometheus {
		g.GET("metrics", gin.WrapH(promhttp.Handler()))
		prometheus.Unregister(collectors.NewGoCollector())
		collectorOpt := collectors.ProcessCollectorOpts(prometheus.ProcessCollectorOpts{})
		collector := collectors.NewProcessCollector(collectorOpt)
		prometheus.Unregister(collector)
	}
	if srv.profile {
		ginPprof.Register(g)
	}
	if srv.tracing {
		g.Use(otelgin.Middleware(config.Conf.Application.Name))
	}
	if len(srv.handlersChain) > 0 {
		g.Use(srv.handlersChain...)
	}
	g.Use(GinZapWithConfig(&GinLoggerConfig{
		TimeFormat: time.RFC3339,
		UTC:        false,
		SkipPaths:  []string{"/healthC"},
	}))
	g.Use(RecoveryWithZap(true))
	g.GET("/healthC", func(c *gin.Context) {
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	})
	srv.gin = g
	return srv
}

// Server gin
func (s *Server) Server() *gin.Engine {
	return s.gin
}

func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.address,
		Handler: s.gin,
	}
	s.httpServer = srv
	logger.Infof("[HTTP] server listening on: %s", s.address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	logger.Info("[HTTP] server stopping")
	return s.httpServer.Shutdown(ctx)
}
