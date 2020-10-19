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

	blockDb     store.Blocks
	validatorDb store.Validators
}

func NewPurgeCmdHandler(cfg *config.Config, blockDb store.Blocks, validatorDb store.Validators) *PurgeCmdHandler {
	return &PurgeCmdHandler{
		cfg: cfg,

		blockDb:     blockDb,
		validatorDb: validatorDb,
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
		return NewPurgeUseCase(h.cfg, h.blockDb, h.validatorDb)
	}
	return h.useCase
}
