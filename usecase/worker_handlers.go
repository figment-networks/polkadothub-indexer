package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewWorkerHandlers(cfg *config.Config, db store.Store, c *client.Client) *WorkerHandlers {
	return &WorkerHandlers{
		RunIndexer:       indexing.NewRunWorkerHandler(cfg, db, c),
		SummarizeIndexer: indexing.NewSummarizeWorkerHandler(cfg, db, c),
		PurgeIndexer:     indexing.NewPurgeWorkerHandler(cfg, db, c),
	}
}

type WorkerHandlers struct {
	RunIndexer       types.WorkerHandler
	SummarizeIndexer types.WorkerHandler
	PurgeIndexer     types.WorkerHandler
}
