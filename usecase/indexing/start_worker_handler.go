package indexing

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

var (
	_ types.WorkerHandler = (*runWorkerHandler)(nil)
)

type runWorkerHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *startUseCase

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

func NewRunWorkerHandler(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *runWorkerHandler {
	return &runWorkerHandler{
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

func (h *runWorkerHandler) Handle() {
	batchSize := h.cfg.DefaultBatchSize
	ctx := context.Background()

	logger.Info(fmt.Sprintf("running indexer use case [handler=worker] [batchSize=%d]", batchSize))

	err := h.getUseCase().Execute(ctx, batchSize)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *runWorkerHandler) getUseCase() *startUseCase {
	if h.useCase == nil {
		return NewStartUseCase(h.cfg, h.client, h.accountEraSeqDb, h.blockSeqDb, h.blockSummaryDb, h.databaseDb, h.eventSeqDb, h.reportsDb,
			h.syncablesDb, h.validatorAggDb, h.validatorEraSeqDb, h.validatorSessionSeqDb, h.validatorSummaryDb,
		)
	}
	return h.useCase
}
