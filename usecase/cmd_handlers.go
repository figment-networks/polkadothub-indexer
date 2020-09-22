package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewCmdHandlers(cfg *config.Config, c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	databaseDb store.Database, eventSeqDb store.EventSeq, reportsDb store.Reports, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *CmdHandlers {
	return &CmdHandlers{
		GetStatus: chain.NewGetStatusCmdHandler(c, syncablesDb),
		StartIndexer: indexing.NewStartCmdHandler(cfg, c, accountEraSeqDb, blockSeqDb, blockSummaryDb, databaseDb, eventSeqDb, reportsDb,
			syncablesDb, validatorAggDb, validatorEraSeqDb, validatorSessionSeqDb, validatorSummaryDb,
		),
		BackfillIndexer: indexing.NewBackfillCmdHandler(cfg, c, accountEraSeqDb, blockSeqDb, blockSummaryDb, databaseDb, eventSeqDb, reportsDb,
			syncablesDb, validatorAggDb, validatorEraSeqDb, validatorSessionSeqDb, validatorSummaryDb,
		),
		PurgeIndexer:     indexing.NewPurgeCmdHandler(cfg, blockSeqDb, blockSummaryDb, validatorSessionSeqDb, validatorSummaryDb),
		SummarizeIndexer: indexing.NewSummarizeCmdHandler(cfg, blockSeqDb, blockSummaryDb, validatorEraSeqDb, validatorSessionSeqDb, validatorSummaryDb),
	}
}

type CmdHandlers struct {
	GetStatus        *chain.GetStatusCmdHandler
	StartIndexer     *indexing.StartCmdHandler
	BackfillIndexer  *indexing.BackfillCmdHandler
	PurgeIndexer     *indexing.PurgeCmdHandler
	SummarizeIndexer *indexing.SummarizeCmdHandler
}
