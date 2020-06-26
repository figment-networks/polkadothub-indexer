package indexing

import (
	"context"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type StartCmdHandler struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	useCase *startUseCase
}

func NewStartCmdHandler(cfg *config.Config, db *store.Store, c *client.Client) *StartCmdHandler {
	return &StartCmdHandler{
		cfg:    cfg,
		db:     db,
		client: c,
	}
}

func (h *StartCmdHandler) Handle(ctx context.Context, batchSize int64) {
	logger.Info(fmt.Sprintf("running indexer use case [handler=cmd] [batchSize=%d]", batchSize))

	err := h.getUseCase().Execute(ctx, batchSize)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *StartCmdHandler) getUseCase() *startUseCase {
	if h.useCase == nil {
		return NewStartUseCase(h.cfg, h.db, h.client)
	}
	return h.useCase
}
