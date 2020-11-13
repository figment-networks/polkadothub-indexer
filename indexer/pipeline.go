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
	CtxReport     = "context_report"
	maxRetries    = 3
	StageAnalyzer = "AnalyzerStage"
)

var (
	ErrIsPristine          = errors.New("cannot run because database is empty")
	ErrIndexCannotBeRun    = errors.New("cannot run index process")
	ErrBackfillCannotBeRun = errors.New("cannot run backfill process")
)

type indexingPipeline struct {
	cfg    *config.Config
	client *client.Client

	status       *pipelineStatus
	configParser ConfigParser
	pipeline     pipeline.CustomPipeline

	databaseDb store.Database
	reportDb   store.Reports
	syncableDb store.Syncables
}

func NewPipeline(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) (*indexingPipeline, error) {
	p := pipeline.NewCustom(NewPayloadFactory())

	// Setup logger
	p.SetLogger(NewLogger())

	// Fetcher stage
	p.AddStage(
		pipeline.NewStageWithTasks(
			pipeline.StageFetcher,
			pipeline.RetryingTask(NewFetcherTask(cli.Height), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorFetcherTask(cli.Validator), isTransient, maxRetries),
		),
	)

	// Syncer stage
	p.AddStage(
		pipeline.NewStageWithTasks(pipeline.StageSyncer, pipeline.RetryingTask(NewMainSyncerTask(syncableDb), isTransient, maxRetries)),
	)

	// Set parser stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StageParser,
			NewBlockParserTask(),
			NewValidatorsParserTask(cli.Account),
		),
	)

	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StageSequencer,
			pipeline.RetryingTask(NewBlockSeqCreatorTask(blockDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSeqCreatorTask(validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSessionSeqCreatorTask(cfg, syncableDb, validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorEraSeqCreatorTask(cfg, syncableDb, validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewEventSeqCreatorTask(eventDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewAccountEraSeqCreatorTask(cfg, accountDb, syncableDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewTransactionSeqCreatorTask(transactionDb), isTransient, maxRetries),
		),
	)

	// Set aggregator stage
	p.AddStage(
		pipeline.NewStageWithTasks(
			pipeline.StageAggregator,
			pipeline.RetryingTask(NewValidatorAggCreatorTask(validatorDb), isTransient, maxRetries),
		),
	)

	// Set analyzer stage
	p.AddStage(
		pipeline.NewStageWithTasks(
			StageAnalyzer,
			pipeline.RetryingTask(NewSystemEventCreatorTask(cfg, accountDb, syncableDb, systemEventDb, validatorDb, validatorDb), isTransient, maxRetries),
		),
	)

	// Set persistor stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StagePersistor,
			pipeline.RetryingTask(NewSyncerPersistorTask(syncableDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewBlockSeqPersistorTask(blockDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSeqPersistorTask(validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorSessionSeqPersistorTask(validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorEraSeqPersistorTask(validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewValidatorAggPersistorTask(validatorDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewEventSeqPersistorTask(eventDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewAccountEraSeqPersistorTask(accountDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewTransactionSeqPersistorTask(transactionDb), isTransient, maxRetries),
			pipeline.RetryingTask(NewSystemEventPersistorTask(systemEventDb), isTransient, 3),
		),
	)

	// Create config parser
	configParser, err := NewConfigParser(cfg.IndexerConfigFile)
	if err != nil {
		return nil, err
	}

	statusChecker := pipelineStatusChecker{syncableDb, configParser.GetCurrentVersionId()}
	pipelineStatus, err := statusChecker.getStatus()
	if err != nil {
		return nil, err
	}

	return &indexingPipeline{
		cfg:    cfg,
		client: cli,

		pipeline:     p,
		status:       pipelineStatus,
		configParser: configParser,

		databaseDb: databaseDb,
		reportDb:   reportDb,
		syncableDb: syncableDb,
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

	source, err := NewIndexSource(p.cfg, p.syncableDb, p.client, &IndexSourceConfig{
		BatchSize:   indexCfg.BatchSize,
		StartHeight: indexCfg.StartHeight,
	})
	if err != nil {
		return err
	}

	sink := NewSink(p.databaseDb, p.syncableDb, indexVersion)

	reportCreator := &reportCreator{
		kind:         model.ReportKindIndex,
		indexVersion: indexVersion,
		startHeight:  source.startHeight,
		endHeight:    source.endHeight,
		reportDb:     p.reportDb,
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
	isLastInSession := p.configParser.IsLastInSession()
	isLastInEra := p.configParser.IsLastInEra()
	source, err := NewBackfillSource(p.cfg, p.syncableDb, p.client, indexVersion, isLastInSession, isLastInEra)
	if err != nil {
		return err
	}

	sink := NewSink(p.databaseDb, p.syncableDb, indexVersion)

	kind := model.ReportKindSequentialReindex
	if backfillCfg.Parallel {
		kind = model.ReportKindParallelReindex
	}

	if backfillCfg.Force {
		if err := p.reportDb.DeleteByKinds([]model.ReportKind{model.ReportKindParallelReindex, model.ReportKindSequentialReindex}); err != nil {
			return err
		}
	}

	reportCreator := &reportCreator{
		kind:         kind,
		indexVersion: indexVersion,
		startHeight:  source.startHeight,
		endHeight:    source.endHeight,
		reportDb:     p.reportDb,
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

	if err := reportCreator.createIfNotExists(model.ReportKindSequentialReindex, model.ReportKindParallelReindex); err != nil {
		return err
	}

	if err := p.syncableDb.SetProcessedAtForRange(reportCreator.report.ID, source.startHeight, source.endHeight); err != nil {
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
