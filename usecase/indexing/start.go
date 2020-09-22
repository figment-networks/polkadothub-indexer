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
	ErrRunningSequentialReindex = errors.New("indexing skipped because sequential reindex hasn't finished yet")
)

type startUseCase struct {
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

func NewStartUseCase(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *startUseCase {
	return &startUseCase{
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

func (uc *startUseCase) Execute(ctx context.Context, batchSize int64) error {
	if err := uc.canExecute(); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountEraSeqDb, uc.blockSeqDb, uc.blockSummaryDb, uc.databaseDb, uc.eventSeqDb, uc.reportsDb,
		uc.syncablesDb, uc.validatorAggDb, uc.validatorEraSeqDb, uc.validatorSessionSeqDb, uc.validatorSummaryDb,
	)
	if err != nil {
		return err
	}

	return indexingPipeline.Start(ctx, indexer.IndexConfig{
		BatchSize: batchSize,
	})
}

// canExecute checks if sequential reindex is already running
// if is it running we skip indexing
func (uc *startUseCase) canExecute() error {
	if _, err := uc.reportsDb.FindNotCompletedByKind(model.ReportKindSequentialReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	return ErrRunningSequentialReindex
}
