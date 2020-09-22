package indexing

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type StartCmdHandler struct {
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

func NewStartCmdHandler(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *StartCmdHandler {
	return &StartCmdHandler{
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

func (h *StartCmdHandler) Handle(ctx context.Context, batchSize int64) {
	logger.Info(fmt.Sprintf("running indexer use case [handler=cmd] [batchSize=%d]", batchSize))

	err := h.getUseCase().Execute(ctx, batchSize)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *StartCmdHandler) getUseCase() *startUseCase {
	if h.useCase == nil {
		return NewStartUseCase(h.cfg, h.client, h.accountEraSeqDb, h.blockSeqDb, h.blockSummaryDb, h.databaseDb, h.eventSeqDb, h.reportsDb,
			h.syncablesDb, h.validatorAggDb, h.validatorEraSeqDb, h.validatorSessionSeqDb, h.validatorSummaryDb,
		)
	}
	return h.useCase
}
