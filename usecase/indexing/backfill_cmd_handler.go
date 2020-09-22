package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type BackfillCmdHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *backfillUseCase

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

func NewBackfillCmdHandler(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *BackfillCmdHandler {
	return &BackfillCmdHandler{
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

func (h *BackfillCmdHandler) Handle(ctx context.Context, parallel bool, force bool, targetIds []int64) {
	logger.Info("running backfill use case [handler=cmd]")

	useCaseConfig := BackfillUseCaseConfig{
		Parallel:  parallel,
		Force:     force,
		TargetIds: targetIds,
	}
	err := h.getUseCase().Execute(ctx, useCaseConfig)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *BackfillCmdHandler) getUseCase() *backfillUseCase {
	if h.useCase == nil {
		return NewBackfillUseCase(h.cfg, h.client, h.accountEraSeqDb, h.blockSeqDb, h.blockSummaryDb, h.databaseDb, h.eventSeqDb, h.reportsDb,
			h.syncablesDb, h.validatorAggDb, h.validatorEraSeqDb, h.validatorSessionSeqDb, h.validatorSummaryDb,
		)
	}
	return h.useCase
}
