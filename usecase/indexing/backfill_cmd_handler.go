package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type BackfillCmdHandler struct {
	cfg    *config.Config
	db     store.Store
	client *client.Client

	useCase *backfillUseCase
}

func NewBackfillCmdHandler(cfg *config.Config, db store.Store, c *client.Client) *BackfillCmdHandler {
	return &BackfillCmdHandler{
		cfg:    cfg,
		db:     db,
		client: c,
	}
}

func (h *BackfillCmdHandler) Handle(ctx context.Context, parallel bool, force bool, targetIds []int64) {
	logger.Info("running backfill use case [handler=cmd]")

	useCaseConfig := BackfillUseCaseConfig{
		Parallel:  parallel,
		Force:     force,
		TargetIds: targetIds,
	}
	err := h.getUseCase().Execute(ctx, useCaseConfig)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *BackfillCmdHandler) getUseCase() *backfillUseCase {
	if h.useCase == nil {
		return NewBackfillUseCase(h.cfg, h.db, h.client)
	}
	return h.useCase
}
