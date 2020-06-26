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
	_ types.WorkerHandler = (*summarizeWorkerHandler)(nil)
)

type summarizeWorkerHandler struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	useCase *summarizeUseCase
}

func NewSummarizeWorkerHandler(cfg *config.Config, db *store.Store, c *client.Client) *summarizeWorkerHandler {
	return &summarizeWorkerHandler{
		cfg:    cfg,
		db:     db,
		client: c,
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
		return NewSummarizeUseCase(h.cfg, h.db)
	}
	return h.useCase
}
