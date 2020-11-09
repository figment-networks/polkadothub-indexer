package server

import (
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/gin-gonic/gin"
	"time"
)

// setupMiddleware sets up middleware for gin application
func (s *Server) setupMiddleware() {
	s.engine.Use(MetricMiddleware())
	s.engine.Use(gin.RecoveryWithWriter(s.writer))
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
