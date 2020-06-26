package indexing

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

var (
	_ types.WorkerHandler = (*purgeWorkerHandler)(nil)
)

type purgeWorkerHandler struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	useCase *purgeUseCase
}

func NewPurgeWorkerHandler(cfg *config.Config, db *store.Store, c *client.Client) *purgeWorkerHandler {
	return &purgeWorkerHandler{
		cfg:    cfg,
		db:     db,
		client: c,
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
		return NewPurgeUseCase(h.cfg, h.db)
	}
	return h.useCase
}



