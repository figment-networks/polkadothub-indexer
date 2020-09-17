package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

func NewCmdHandlers(cfg *config.Config, db store.Store, c *client.Client) *CmdHandlers {
	return &CmdHandlers{
		GetStatus:        chain.NewGetStatusCmdHandler(db, c),
		StartIndexer:     indexing.NewStartCmdHandler(cfg, db, c),
		BackfillIndexer:  indexing.NewBackfillCmdHandler(cfg, db, c),
		PurgeIndexer:     indexing.NewPurgeCmdHandler(cfg, db, c),
		SummarizeIndexer: indexing.NewSummarizeCmdHandler(cfg, db, c),
	}
}

type CmdHandlers struct {
	GetStatus        *chain.GetStatusCmdHandler
	StartIndexer     *indexing.StartCmdHandler
	BackfillIndexer  *indexing.BackfillCmdHandler
	PurgeIndexer     *indexing.PurgeCmdHandler
	SummarizeIndexer *indexing.SummarizeCmdHandler
}
