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
	CtxReport = "context_report"
)

type indexingPipeline struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	targetsReader *targetsReader
	pipeline      pipeline.CustomPipeline
}

func NewPipeline(cfg *config.Config, db *store.Store, client *client.Client) (*indexingPipeline, error) {
	p := pipeline.NewCustom(NewPayloadFactory())

	// Setup logger
	p.SetLogger(NewLogger())

	// Fetcher stage
	p.AddStage(
		pipeline.NewStageWithTasks(pipeline.StageFetcher, pipeline.RetryingTask(NewFetcherTask(client.Height), isTransient, 3)),
	)

	// Syncer stage
	p.AddStage(
		pipeline.NewStageWithTasks(pipeline.StageSyncer, pipeline.RetryingTask(NewMainSyncerTask(db), isTransient, 3)),
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
			pipeline.RetryingTask(NewBlockSeqCreatorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewValidatorSessionSeqCreatorTask(cfg, db), isTransient, 3),
			pipeline.RetryingTask(NewValidatorEraSeqCreatorTask(cfg, db), isTransient, 3),
			pipeline.RetryingTask(NewEventSeqCreatorTask(db), isTransient, 3),
		),
	)

	// Set aggregator stage
	p.AddStage(
		pipeline.NewStageWithTasks(
			pipeline.StageAggregator,
			pipeline.RetryingTask(NewValidatorAggCreatorTask(db), isTransient, 3),
		),
	)

	// Set persistor stage
	p.AddStage(
		pipeline.NewAsyncStageWithTasks(
			pipeline.StagePersistor,
			pipeline.RetryingTask(NewSyncerPersistorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewBlockSeqPersistorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewValidatorSessionSeqPersistorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewValidatorEraSeqPersistorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewValidatorAggPersistorTask(db), isTransient, 3),
			pipeline.RetryingTask(NewEventSeqPersistorTask(db), isTransient, 3),
		),
	)

	// Create targets reader
	targetsReader, err := NewTargetsReader(cfg.IndexerTargetsFile)
	if err != nil {
		return nil, err
	}

	return &indexingPipeline{
		cfg:    cfg,
		db:     db,
		client: client,

		pipeline:      p,
		targetsReader: targetsReader,
	}, nil
}

type StartConfig struct {
	BatchSize   int64
	StartHeight int64
}

func (p *indexingPipeline) Start(ctx context.Context, startCfg StartConfig) error {
	indexVersion := p.targetsReader.GetCurrentVersion()

	source, err := NewIndexSource(p.cfg, p.db, p.client, &IndexSourceConfig{
		BatchSize:   startCfg.BatchSize,
		StartHeight: startCfg.StartHeight,
	})
	if err != nil {
		return err
	}

	sink := NewSink(p.db, indexVersion)

	report, err := p.createReport(source.startHeight, source.endHeight, model.ReportKindIndex, indexVersion)
	if err != nil {
		return err
	}

	ctxWithReport := context.WithValue(ctx, CtxReport, report)

	pipelineOptions, err := p.getPipelineOptions(false)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("starting pipeline [start=%d] [end=%d]", source.startHeight, source.endHeight))

	err = p.pipeline.Start(ctxWithReport, source, sink, pipelineOptions)
	if err != nil {
		metric.IndexerTotalErrors.Inc()
	}

	logger.Info(fmt.Sprintf("pipeline completed [Err: %+v]", err))

	err = p.completeReport(report, source.Len(), sink.successCount, err)

	return err
}

type BackfillConfig struct {
	Parallel  bool
	Force     bool
	TargetIds []int64
}

func (p *indexingPipeline) Backfill(ctx context.Context, backfillCfg BackfillConfig) error {
	indexVersion := p.targetsReader.GetCurrentVersion()

	source, err := NewBackfillSource(p.cfg, p.db, p.client, &BackfillSourceConfig{
		indexVersion: indexVersion,
	})
	if err != nil {
		return err
	}

	sink := NewSink(p.db, indexVersion)

	kind := model.ReportKindSequentialReindex
	if backfillCfg.Parallel {
		kind = model.ReportKindParallelReindex
	}

	if backfillCfg.Force {
		if err := p.db.Reports.DeleteReindexing(); err != nil {
			return err
		}
	}

	report, err := p.getReport(indexVersion, source, kind)
	if err != nil {
		return err
	}

	if err := p.db.Syncables.SetProcessedAtForRange(report.ID, source.startHeight, source.endHeight); err != nil {
		return err
	}

	ctxWithReport := context.WithValue(ctx, CtxReport, report)

	pipelineOptions, err := p.getPipelineOptions(false, backfillCfg.TargetIds...)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("starting pipeline backfill [start=%d] [end=%d] [kind=%s]", source.startHeight, source.endHeight, kind))

	if err := p.pipeline.Start(ctxWithReport, source, sink, pipelineOptions); err != nil {
		return err
	}

	if err = p.completeReport(report, source.Len(), sink.successCount, err); err != nil {
		return err
	}

	logger.Info("pipeline backfill completed")

	return nil
}

type RunConfig struct {
	Height          int64
	DesiredTargetID int64
	Dry             bool
}

func (p *indexingPipeline) Run(ctx context.Context, runCfg RunConfig) (*payload, error) {
	pipelineOptions, err := p.getPipelineOptions(runCfg.Dry, runCfg.DesiredTargetID)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("running pipeline... [height=%d] [version=%d]", runCfg.Height, runCfg.DesiredTargetID))

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

func (p *indexingPipeline) getPipelineOptions(dry bool, targetIds ...int64) (*pipeline.Options, error) {
	taskWhitelist, err := p.getTasksWhitelist(targetIds)
	if err != nil {
		return nil, err
	}

	return &pipeline.Options{
		TaskWhitelist:   taskWhitelist,
		StagesBlacklist: p.getStagesBlacklist(dry),
	}, nil
}

func (p *indexingPipeline) getTasksWhitelist(targetIds []int64) ([]pipeline.TaskName, error) {
	var taskWhitelist []pipeline.TaskName
	var err error
	if len(targetIds) == 0 {
		taskWhitelist = p.targetsReader.GetAllTasks()
	} else if len(targetIds) == 1 {
		taskWhitelist, err = p.targetsReader.GetTasksByTargetId(targetIds[0])
	} else {
		taskWhitelist, err = p.targetsReader.GetTasksByTargetIds(targetIds)
	}
	return taskWhitelist, err
}

func (p *indexingPipeline) getStagesBlacklist(dry bool) []pipeline.StageName {
	var stagesBlacklist []pipeline.StageName
	if dry {
		stagesBlacklist = append(stagesBlacklist, pipeline.StagePersistor)
	}
	return stagesBlacklist
}

func (p *indexingPipeline) getReport(indexVersion int64, source *backfillSource, kind model.ReportKind) (*model.Report, error) {
	report, err := p.db.Reports.FindNotCompletedByIndexVersion(indexVersion, model.ReportKindSequentialReindex, model.ReportKindParallelReindex)
	if err != nil {
		if err == store.ErrNotFound {
			report, err = p.createReport(source.startHeight, source.endHeight, kind, indexVersion)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if report.Kind != kind {
			return nil, errors.New(fmt.Sprintf("there is already reindexing in process [kind=%s] (use -force flag to override it)", report.Kind))
		}
	}
	return report, nil
}

func (p *indexingPipeline) createReport(startHeight int64, endHeight int64, kind model.ReportKind, indexVersion int64) (*model.Report, error) {
	report := &model.Report{
		Kind:         kind,
		IndexVersion: indexVersion,
		StartHeight:  startHeight,
		EndHeight:    endHeight,
	}
	if err := p.db.Reports.Create(report); err != nil {
		return nil, err
	}
	return report, nil
}

func (p *indexingPipeline) completeReport(report *model.Report, totalCount int64, successCount int64, err error) error {
	report.Complete(successCount, totalCount-successCount, err)

	return p.db.Reports.Save(report)
}

func isTransient(error) bool {
	return true
}
