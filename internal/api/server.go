package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zxCroshka/wb-test/internal/aggregator"
	"github.com/zxCroshka/wb-test/internal/metrics"
)

type Server struct {
	aggregator *aggregator.Aggregator
	metrics    *metrics.Metrics
	engine     *gin.Engine
	httpServer *http.Server
}

func NewServer(
	agg *aggregator.Aggregator,
	m *metrics.Metrics,
) *Server {

	r := gin.Default()

	r.Use(metricsMiddleware(m))

	server := &Server{
		aggregator: agg,
		metrics:    m,
		engine:     r,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	s.engine.GET("/health", s.handleHealth)
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.engine.GET("/top", s.handleTop)

	stoplistGroup := s.engine.Group("/stoplist")
	{
		stoplistGroup.GET("", s.handleStopListGetAll)
		stoplistGroup.POST("", s.handleStopListAdd)
		stoplistGroup.DELETE("", s.handleStopListRemove)
		stoplistGroup.GET("/check", s.handleStopListCheck)
		stoplistGroup.POST("/batch", s.handleStopListBatchAdd)
		stoplistGroup.DELETE("/all", s.handleStopListClear)
	}
}

func (s *Server) Start(port string, readTimeout, writeTimeout, idleTimeout time.Duration) error {
	s.httpServer = &http.Server{
		Addr:         port,
		Handler:      s.engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
