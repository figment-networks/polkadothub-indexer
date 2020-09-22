package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

var (
	_ types.WorkerHandler = (*summarizeWorkerHandler)(nil)
)

type summarizeWorkerHandler struct {
	cfg *config.Config

	useCase *summarizeUseCase

	blockSeqDb            store.BlockSeq
	blockSummaryDb        store.BlockSummary
	validatorEraSeqDb     store.ValidatorEraSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewSummarizeWorkerHandler(cfg *config.Config, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary, validatorEraSeqDb store.ValidatorEraSeq,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *summarizeWorkerHandler {
	return &summarizeWorkerHandler{
		cfg:                   cfg,
		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		validatorEraSeqDb:     validatorEraSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

func (h *summarizeWorkerHandler) Handle() {
	ctx := context.Background()

	logger.Info("running summarize use case [handler=worker]")

	err := h.getUseCase().Execute(ctx)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *summarizeWorkerHandler) getUseCase() *summarizeUseCase {
	if h.useCase == nil {
		return NewSummarizeUseCase(h.cfg, h.blockSeqDb, h.blockSummaryDb, h.validatorEraSeqDb, h.validatorSessionSeqDb, h.validatorSummaryDb)
	}
	return h.useCase
}
