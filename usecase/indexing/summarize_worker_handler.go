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

	blockDb     store.Blocks
	validatorDb store.Validators
}

func NewSummarizeWorkerHandler(cfg *config.Config, blockDb store.Blocks, validatorDb store.Validators) *summarizeWorkerHandler {
	return &summarizeWorkerHandler{
		cfg:         cfg,
		blockDb:     blockDb,
		validatorDb: validatorDb,
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
		return NewSummarizeUseCase(h.cfg, h.blockDb, h.validatorDb)
	}
	return h.useCase
}
