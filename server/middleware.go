package server

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/utils/reporting"
	"github.com/gin-gonic/gin"
)

// setupMiddleware sets up middleware for gin application
func (s *Server) setupMiddleware() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(MetricMiddleware())
	s.engine.Use(ErrorReportingMiddleware())
}

// MetricMiddleware is a middleware responsible for logging query execution time metric
func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		elapsed := time.Since(t)

		metric.ServerRequestDuration.WithLabelValues(c.Request.URL.Path).Set(elapsed.Seconds())
	}
}

func ErrorReportingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer reporting.RecoverError()
		c.Next()
	}
}
