package metric

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	DatabaseQueryDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "figment",
			Subsystem: "database",
			Name: "query_duration",
			Help: "The total time required to execute query on database",
		},
		[]string{"query"},
	)

	ServerRequestDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "figment",
			Subsystem: "server",
			Name: "request_duration",
			Help: "The total time required to execute http request",
		},
		[]string{"request"},
	)
)

// ServerMetric handles HTTP requests
type ServerMetric struct {}

// NewDatabaseMetric returns a new server instance
func NewServerMetric() *ServerMetric {
	app := &ServerMetric{}
	return app.init()
}

func (m *ServerMetric) StartServer(listenAdd string, url string) error {
	logger.Info(fmt.Sprintf("starting metric server at %s...", url), logger.Field("app", "server"))

	http.Handle(url, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	return http.ListenAndServe(listenAdd, nil)
}

func (m *ServerMetric) init() *ServerMetric {
	logger.Info("initializing metric server...", logger.Field("app", "server"))

	prometheus.MustRegister(DatabaseQueryDuration)
	prometheus.MustRegister(ServerRequestDuration)

	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
	return m
}