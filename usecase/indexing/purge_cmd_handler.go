package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type PurgeCmdHandler struct {
	cfg *config.Config

	useCase *purgeUseCase

	blockSeqDb            store.BlockSeq
	blockSummaryDb        store.BlockSummary
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewPurgeCmdHandler(cfg *config.Config, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *PurgeCmdHandler {
	return &PurgeCmdHandler{
		cfg: cfg,

		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

func (h *PurgeCmdHandler) Handle(ctx context.Context) {
	logger.Info("running purge use case [handler=cmd]")

	err := h.getUseCase().Execute(ctx)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *PurgeCmdHandler) getUseCase() *purgeUseCase {
	if h.useCase == nil {
		return NewPurgeUseCase(h.cfg, h.blockSeqDb, h.blockSummaryDb, h.validatorSessionSeqDb, h.validatorSummaryDb)
	}
	return h.useCase
}
