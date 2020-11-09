package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewCmdHandlers(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, transactionDb store.Transactions, validatorDb store.Validators,
) *CmdHandlers {
	return &CmdHandlers{
		GetStatus:        chain.NewGetStatusCmdHandler(cli, syncableDb),
		StartIndexer:     indexing.NewStartCmdHandler(cfg, cli, accountDb, blockDb, databaseDb, eventDb, reportDb, syncableDb, transactionDb, validatorDb),
		BackfillIndexer:  indexing.NewBackfillCmdHandler(cfg, cli, accountDb, blockDb, databaseDb, eventDb, reportDb, syncableDb, transactionDb, validatorDb),
		PurgeIndexer:     indexing.NewPurgeCmdHandler(cfg, blockDb, validatorDb),
		SummarizeIndexer: indexing.NewSummarizeCmdHandler(cfg, blockDb, validatorDb),
	}
}

type CmdHandlers struct {
	GetStatus        *chain.GetStatusCmdHandler
	StartIndexer     *indexing.StartCmdHandler
	BackfillIndexer  *indexing.BackfillCmdHandler
	PurgeIndexer     *indexing.PurgeCmdHandler
	SummarizeIndexer *indexing.SummarizeCmdHandler
}
