package metric

import (
	"fmt"
	"net/http"
	"time"

	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	IndexerUseCaseDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "figment",
			Subsystem: "indexer",
			Name:      "use_case_duration",
			Help:      "The total time required to execute use case",
		},
		[]string{"task"},
	)

	IndexerDbSizeAfterHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "figment",
		Subsystem: "indexer",
		Name:      "db_size",
		Help:      "The size of the database after indexing of height",
	})
)

// IndexerMetric handles HTTP requests
type IndexerMetric struct{}

// NewIndexerMetric returns a new server instance
func NewIndexerMetric() *IndexerMetric {
	app := &IndexerMetric{}
	return app.init()
}

func (m *IndexerMetric) StartServer(listenAdd string, url string) error {
	logger.Info(fmt.Sprintf("starting metric server at %s...", url), logger.Field("app", "indexer"))

	http.Handle(url, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	return http.ListenAndServe(listenAdd, nil)
}

func (m *IndexerMetric) init() *IndexerMetric {
	logger.Info("initializing metric server...", logger.Field("app", "indexer"))

	prometheus.MustRegister(IndexerUseCaseDuration)
	prometheus.MustRegister(IndexerDbSizeAfterHeight)

	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
	return m
}

func LogUseCaseDuration(start time.Time, useCaseName string) {
	elapsed := time.Since(start)
	IndexerUseCaseDuration.WithLabelValues(useCaseName).Set(elapsed.Seconds())
}
