package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

var (
	ErrBackfillRunning = errors.New("backfill already running (use force flag to override it)")
)

type backfillUseCase struct {
	cfg    *config.Config
	client *client.Client

	accountEraSeqDb       store.AccountEraSeq
	blockSeqDb            store.BlockSeq
	blockSummaryDb        store.BlockSummary
	databaseDb            store.Database
	eventSeqDb            store.EventSeq
	reportsDb             store.Reports
	syncablesDb           store.Syncables
	validatorAggDb        store.ValidatorAgg
	validatorEraSeqDb     store.ValidatorEraSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewBackfillUseCase(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *backfillUseCase {
	return &backfillUseCase{
		cfg:    cfg,
		client: c,

		accountEraSeqDb:       accountEraSeqDb,
		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		databaseDb:            databaseDb,
		eventSeqDb:            eventSeqDb,
		reportsDb:             reportsDb,
		syncablesDb:           syncablesDb,
		validatorAggDb:        validatorAggDb,
		validatorEraSeqDb:     validatorEraSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

type BackfillUseCaseConfig struct {
	Parallel  bool
	Force     bool
	TargetIds []int64
}

func (uc *backfillUseCase) Execute(ctx context.Context, useCaseConfig BackfillUseCaseConfig) error {
	if err := uc.canExecute(); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountEraSeqDb, uc.blockSeqDb, uc.blockSummaryDb, uc.databaseDb, uc.eventSeqDb, uc.reportsDb,
		uc.syncablesDb, uc.validatorAggDb, uc.validatorEraSeqDb, uc.validatorSessionSeqDb, uc.validatorSummaryDb,
	)
	if err != nil {
		return err
	}

	return indexingPipeline.Backfill(ctx, indexer.BackfillConfig{
		Parallel:  useCaseConfig.Parallel,
		Force:     useCaseConfig.Force,
		TargetIds: useCaseConfig.TargetIds,
	})
}

// canExecute checks if reindex is already running
// if is it running we skip indexing
func (uc *backfillUseCase) canExecute() error {
	if _, err := uc.reportsDb.FindNotCompletedByKind(model.ReportKindSequentialReindex, model.ReportKindParallelReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err

	}
	return ErrBackfillRunning
}
