package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewWorkerHandlers(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *WorkerHandlers {
	return &WorkerHandlers{
		RunIndexer: indexing.NewRunWorkerHandler(cfg, c, accountEraSeqDb, blockSeqDb, blockSummaryDb, databaseDb, eventSeqDb, reportsDb,
			syncablesDb, validatorAggDb, validatorEraSeqDb, validatorSessionSeqDb, validatorSummaryDb,
		),
		SummarizeIndexer: indexing.NewSummarizeWorkerHandler(cfg, blockSeqDb, blockSummaryDb, validatorEraSeqDb, validatorSessionSeqDb, validatorSummaryDb),
		PurgeIndexer:     indexing.NewPurgeWorkerHandler(cfg, blockSeqDb, blockSummaryDb, validatorSessionSeqDb, validatorSummaryDb),
	}
}

type WorkerHandlers struct {
	RunIndexer       types.WorkerHandler
	SummarizeIndexer types.WorkerHandler
	PurgeIndexer     types.WorkerHandler
}
