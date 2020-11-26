package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewWorkerHandlers(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *WorkerHandlers {
	return &WorkerHandlers{
		RunIndexer:       indexing.NewRunWorkerHandler(cfg, cli, accountDb, blockDb, databaseDb, eventDb, reportDb, rewardDb, syncableDb, systemEventDb, transactionDb, validatorDb),
		SummarizeIndexer: indexing.NewSummarizeWorkerHandler(cfg, blockDb, validatorDb),
		PurgeIndexer:     indexing.NewPurgeWorkerHandler(cfg, blockDb, validatorDb),
	}
}

type WorkerHandlers struct {
	RunIndexer       types.WorkerHandler
	SummarizeIndexer types.WorkerHandler
	PurgeIndexer     types.WorkerHandler
}
