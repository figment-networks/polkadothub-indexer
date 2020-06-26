package worker

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/usecase"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/figment-networks/polkadothub-indexer/utils/reporting"
	"github.com/robfig/cron/v3"
)

var (
	job cron.Job
)

type Worker struct {
	cfg      *config.Config
	handlers *usecase.WorkerHandlers

	logger  logger.CronLogger
	cronJob *cron.Cron
}

func New(cfg *config.Config, handlers *usecase.WorkerHandlers) (*Worker, error) {
	log := logger.NewCronLogger()
	cronJob := cron.New(
		cron.WithLogger(cron.VerbosePrintfLogger(log)),
		cron.WithChain(
			cron.Recover(log),
		),
	)

	w := &Worker{
		cfg:      cfg,
		handlers: handlers,
		logger:   log,
		cronJob:  cronJob,
	}

	return w.init()
}

func (w *Worker) init() (*Worker, error) {
	_, err := w.addRunIndexerJob()
	if err != nil {
		return nil, err
	}

	_, err = w.addSummarizeIndexerJob()
	if err != nil {
		return nil, err
	}

	_, err = w.addPurgeIndexerJob()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Worker) Start() error {
	defer reporting.RecoverError()

	logger.Info("starting worker...", logger.Field("app", "worker"))

	w.cronJob.Start()

	return w.startMetricsServer()
}

func (w *Worker) startMetricsServer() error {
	return metric.NewIndexerMetric().StartServer(w.cfg.IndexerMetricAddr, w.cfg.MetricServerUrl)
}
