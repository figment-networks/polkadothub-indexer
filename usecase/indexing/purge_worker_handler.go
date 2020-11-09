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

	blockDb     store.Blocks
	validatorDb store.Validators
}

func NewPurgeWorkerHandler(cfg *config.Config, blockDb store.Blocks, validatorDb store.Validators) *purgeWorkerHandler {
	return &purgeWorkerHandler{
		cfg: cfg,

		blockDb:     blockDb,
		validatorDb: validatorDb,
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
		return NewPurgeUseCase(h.cfg, h.blockDb, h.validatorDb)
	}
	return h.useCase
}
