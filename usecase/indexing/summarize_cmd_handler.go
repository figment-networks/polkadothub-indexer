package indexing

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type SummarizeCmdHandler struct {
	cfg *config.Config

	useCase *summarizeUseCase

	blockSeqDb            store.BlockSeq
	blockSummaryDb        store.BlockSummary
	validatorEraSeqDb     store.ValidatorEraSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewSummarizeCmdHandler(cfg *config.Config, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary, validatorEraSeqDb store.ValidatorEraSeq,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *SummarizeCmdHandler {
	return &SummarizeCmdHandler{
		cfg: cfg,

		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		validatorEraSeqDb:     validatorEraSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

func (h *SummarizeCmdHandler) Handle(ctx context.Context) {
	logger.Info(fmt.Sprintf("summarizing indexer use case [handler=cmd]"))

	err := h.getUseCase().Execute(ctx)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *SummarizeCmdHandler) getUseCase() *summarizeUseCase {
	if h.useCase == nil {
		return NewSummarizeUseCase(h.cfg, h.blockSeqDb, h.blockSummaryDb, h.validatorEraSeqDb, h.validatorSessionSeqDb, h.validatorSummaryDb)
	}
	return h.useCase
}
