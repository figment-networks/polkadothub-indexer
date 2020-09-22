package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

var (
	_ types.WorkerHandler = (*purgeWorkerHandler)(nil)
)

type purgeWorkerHandler struct {
	cfg *config.Config

	useCase *purgeUseCase

	blockSeqDb            store.BlockSeq
	blockSummaryDb        store.BlockSummary
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewPurgeWorkerHandler(cfg *config.Config, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *purgeWorkerHandler {
	return &purgeWorkerHandler{
		cfg: cfg,

		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

func (h *purgeWorkerHandler) Handle() {
	ctx := context.Background()

	logger.Info("running purge use case [handler=worker]")

	err := h.getUseCase().Execute(ctx)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *purgeWorkerHandler) getUseCase() *purgeUseCase {
	if h.useCase == nil {
		return NewPurgeUseCase(h.cfg, h.blockSeqDb, h.blockSummaryDb, h.validatorSessionSeqDb, h.validatorSummaryDb)
	}
	return h.useCase
}
