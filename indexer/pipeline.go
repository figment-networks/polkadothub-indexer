package indexer

import (
	"context"
	"fmt"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/pkg/errors"
)

const (
	CtxReport  = "context_report"
	maxRetries = 3
)

var (
	ErrIsPristine          = errors.New("cannot run because database is empty")
	ErrIndexCannotBeRun    = errors.New("cannot run index process")
	ErrBackfillCannotBeRun = errors.New("cannot run backfill process")
)

type indexingPipeline struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	status       *pipelineStatus
	configParser ConfigParser
	pipeline     pipeline.CustomPipeline
}

func NewPipeline(cfg *config.Config, db *store.Store, client *client.Client) (*indexingPipeline, error) {
	p := pipeline.NewCustom(NewPayloadFactory())

	// Setup logger
	p.SetLogger(NewLogger())

	// Fetcher stage
	p.AddStage(
		pipeline.NewStageWithTasks(pipeline.StageFetcher, pipeline.RetryingTask(NewFetcherTask(client.Height), isTransient, maxRetries)),
	)

	// Syncer stage
	p.AddStage(
		pipeline.NewStageWithTasks(pipeline.StageSyncer, pipeline.RetryingTask(NewMainSyncerTask(db), isTransient, maxRetries)),
	)

	// Set parser stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StageParser,
			NewBlockParserTask(),
			NewValidatorsParserTask(),
		),
	)

	// Set sequencer stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StageSequencer,
			pipeline.RetryingTask(NewBlockSeqCreatorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSessionSeqCreatorTask(cfg, db), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorEraSeqCreatorTask(cfg, db), isTransient, maxRetries),
			pipeline.RetryingTask(NewEventSeqCreatorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewAccountEraSeqCreatorTask(cfg, db), isTransient, maxRetries),
		),
	)

	// Set aggregator stage
	p.AddStage(
		pipeline.NewStageWithTasks(
			pipeline.StageAggregator,
			pipeline.RetryingTask(NewValidatorAggCreatorTask(db), isTransient, maxRetries),
		),
	)

	// Set persistor stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StagePersistor,
			pipeline.RetryingTask(NewSyncerPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewBlockSeqPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSessionSeqPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorEraSeqPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorAggPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewEventSeqPersistorTask(db), isTransient, maxRetries),
			pipeline.RetryingTask(NewAccountEraSeqPersistorTask(db), isTransient, maxRetries),
		),
	)

	// Create config parser
	configParser, err := NewConfigParser(cfg.IndexerConfigFile)
	if err != nil {
		return nil, err
	}

	statusChecker := pipelineStatusChecker{db.Syncables, configParser.GetCurrentVersionId()}
	pipelineStatus, err := statusChecker.getStatus()
	if err != nil {
		return nil, err
	}

	return &indexingPipeline{
		cfg:    cfg,
		db:     db,
		client: client,

		pipeline:     p,
		status:       pipelineStatus,
		configParser: configParser,
	}, nil
}

type IndexConfig struct {
	BatchSize   int64
	StartHeight int64
}

// Start starts indexing process
func (p *indexingPipeline) Start(ctx context.Context, indexCfg IndexConfig) error {
	if err := p.canRunIndex(); err != nil {
		return err
	}

	indexVersion := p.configParser.GetCurrentVersionId()

	source, err := NewIndexSource(p.cfg, p.db, p.client, &IndexSourceConfig{
		BatchSize:   indexCfg.BatchSize,
		StartHeight: indexCfg.StartHeight,
	})
	if err != nil {
		return err
	}

	sink := NewSink(p.db, indexVersion)

	reportCreator := &reportCreator{
		kind:         model.ReportKindIndex,
		indexVersion: indexVersion,
		startHeight:  source.startHeight,
		endHeight:    source.endHeight,
		store:        p.db.Reports,
	}

	if err := reportCreator.create(); err != nil {
		return err
	}

	versionIds := p.configParser.GetAllVersionedVersionIds()
	pipelineOptionsCreator := &pipelineOptionsCreator{
		configParser:      p.configParser,
		desiredVersionIds: versionIds,
	}
	pipelineOptions, err := pipelineOptionsCreator.parse()
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("starting pipeline [start=%d] [end=%d]", source.startHeight, source.endHeight))

	ctxWithReport := context.WithValue(ctx, CtxReport, reportCreator.report)
	err = p.pipeline.Start(ctxWithReport, source, sink, pipelineOptions)
	if err != nil {
		metric.IndexerTotalErrors.Inc()
	}

	logger.Info(fmt.Sprintf("pipeline completed [Err: %+v]", err))

	err = reportCreator.complete(source.Len(), sink.successCount, err)
	return err
}

func (p *indexingPipeline) canRunIndex() error {
	if !p.status.isPristine && !p.status.isUpToDate {
		if p.configParser.IsAnyVersionSequential(p.status.missingVersionIds) {
			return ErrIndexCannotBeRun
		}
	}
	return nil
}

type BackfillConfig struct {
	Parallel  bool
	Force     bool
	TargetIds []int64
}

func (p *indexingPipeline) Backfill(ctx context.Context, backfillCfg BackfillConfig) error {
	if err := p.canRunBackfill(backfillCfg.Parallel); err != nil {
		return err
	}

	indexVersion := p.configParser.GetCurrentVersionId()

	source, err := NewBackfillSource(p.cfg, p.db, p.client, indexVersion)
	if err != nil {
		return err
	}

	sink := NewSink(p.db, indexVersion)

	kind := model.ReportKindSequentialReindex
	if backfillCfg.Parallel {
		kind = model.ReportKindParallelReindex
	}

	if backfillCfg.Force {
		if err := p.db.Reports.DeleteByKinds([]model.ReportKind{model.ReportKindParallelReindex, model.ReportKindSequentialReindex}); err != nil {
			return err
		}
	}

	reportCreator := &reportCreator{
		kind:         kind,
		indexVersion: indexVersion,
		startHeight:  source.startHeight,
		endHeight:    source.endHeight,
		store:        p.db.Reports,
	}

	versionIds := p.status.missingVersionIds
	pipelineOptionsCreator := &pipelineOptionsCreator{
		configParser:      p.configParser,
		desiredVersionIds: versionIds,
	}
	pipelineOptions, err := pipelineOptionsCreator.parse()
	if err != nil {
		return err
	}

	if err := p.db.Syncables.SetProcessedAtForRange(reportCreator.report.ID, source.startHeight, source.endHeight); err != nil {
		return err
	}

	ctxWithReport := context.WithValue(ctx, CtxReport, reportCreator.report)

	logger.Info(fmt.Sprintf("starting pipeline backfill [start=%d] [end=%d] [kind=%s]", source.startHeight, source.endHeight, kind))

	if err := p.pipeline.Start(ctxWithReport, source, sink, pipelineOptions); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("pipeline completed [Err: %+v]", err))

	err = reportCreator.complete(source.Len(), sink.successCount, err)

	return err
}

func (p *indexingPipeline) canRunBackfill(isParallel bool) error {
	if p.status.isPristine {
		return ErrIsPristine
	}

	if !p.status.isUpToDate {
		if isParallel && p.configParser.IsAnyVersionSequential(p.status.missingVersionIds) {
			return ErrBackfillCannotBeRun
		}
	}
	return nil
}

type RunConfig struct {
	Height            int64
	DesiredVersionIDs []int64
	DesiredTargetIDs  []int64
	Dry               bool
}

func (p *indexingPipeline) Run(ctx context.Context, runCfg RunConfig) (*payload, error) {
	pipelineOptionsCreator := &pipelineOptionsCreator{
		configParser:      p.configParser,
		dry:               runCfg.Dry,
		desiredVersionIds: runCfg.DesiredVersionIDs,
		desiredTargetIds:  runCfg.DesiredTargetIDs,
	}

	pipelineOptions, err := pipelineOptionsCreator.parse()
	if err != nil {
		return nil, err
	}

	logger.Info("running pipeline...",
		logger.Field("height", runCfg.Height),
		logger.Field("versions", runCfg.DesiredVersionIDs),
		logger.Field("targets", runCfg.DesiredTargetIDs),
	)

	runPayload, err := p.pipeline.Run(ctx, runCfg.Height, pipelineOptions)
	if err != nil {
		metric.IndexerTotalErrors.Inc()
		logger.Info(fmt.Sprintf("pipeline completed with error [Err: %+v]", err))
		return nil, err
	}

	logger.Info("pipeline completed successfully")

	payload := runPayload.(*payload)
	return payload, nil
}

func isTransient(error) bool {
	return true
}
