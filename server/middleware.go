package server

import (
	"strings"
	"time"

	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/gin-gonic/gin"
)

// setupMiddleware sets up middleware for gin application
func (s *Server) setupMiddleware() {
	//s.engine.Use(MetricMiddleware())
	s.engine.Use(gin.RecoveryWithWriter(s.writer))
}

// MetricMiddleware is a middleware responsible for logging query execution time metric
func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		if path := strings.Split(strings.TrimPrefix(c.Request.URL.Path, "/"), "/"); len(path) > 0 {
			metric.ServerRequestDuration.WithLabelValues(path[0]).Set(time.Since(t).Seconds())
		}
	}
}
