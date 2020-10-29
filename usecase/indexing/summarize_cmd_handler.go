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

	blockDb     store.Blocks
	validatorDb store.Validators
}

func NewSummarizeCmdHandler(cfg *config.Config, blockDb store.Blocks, validatorDb store.Validators) *SummarizeCmdHandler {
	return &SummarizeCmdHandler{
		cfg: cfg,

		blockDb:     blockDb,
		validatorDb: validatorDb,
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
		return NewSummarizeUseCase(h.cfg, h.blockDb, h.validatorDb)
	}
	return h.useCase
}
