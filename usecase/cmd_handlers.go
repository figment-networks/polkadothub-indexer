package usecase

import (
	"fmt"
	"reflect"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

type CmdHandlerParams struct {
	Config                *config.Config
	Client                *client.Client
	AccountEraSeqDb       store.AccountEraSeq
	BlockSeqDb            store.BlockSeq
	BlockSummaryDb        store.BlockSummary
	DatabaseDb            store.Database
	EventSeqDb            store.EventSeq
	ReportsDb             store.Reports
	SyncablesDb           store.Syncables
	ValidatorAggDb        store.ValidatorAgg
	ValidatorEraSeqDb     store.ValidatorEraSeq
	ValidatorSessionSeqDb store.ValidatorSessionSeq
	ValidatorSummaryDb    store.ValidatorSummary
}

func NewCmdHandlers(b *CmdHandlerParams) (*CmdHandlers, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}

	return &CmdHandlers{
		GetStatus: chain.NewGetStatusCmdHandler(b.Client, b.SyncablesDb),
		StartIndexer: indexing.NewStartCmdHandler(b.Config, b.Client, b.AccountEraSeqDb, b.BlockSeqDb, b.BlockSummaryDb, b.DatabaseDb, b.EventSeqDb, b.ReportsDb,
			b.SyncablesDb, b.ValidatorAggDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb,
		),
		BackfillIndexer: indexing.NewBackfillCmdHandler(b.Config, b.Client, b.AccountEraSeqDb, b.BlockSeqDb, b.BlockSummaryDb, b.DatabaseDb, b.EventSeqDb, b.ReportsDb,
			b.SyncablesDb, b.ValidatorAggDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb,
		),
		PurgeIndexer:     indexing.NewPurgeCmdHandler(b.Config, b.BlockSeqDb, b.BlockSummaryDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb),
		SummarizeIndexer: indexing.NewSummarizeCmdHandler(b.Config, b.BlockSeqDb, b.BlockSummaryDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb),
	}, nil
}

type CmdHandlers struct {
	GetStatus        *chain.GetStatusCmdHandler
	StartIndexer     *indexing.StartCmdHandler
	BackfillIndexer  *indexing.BackfillCmdHandler
	PurgeIndexer     *indexing.PurgeCmdHandler
	SummarizeIndexer *indexing.SummarizeCmdHandler
}

func (b *CmdHandlerParams) validate() error {
	if b == nil {
		return fmt.Errorf("CmdHandlerParams cannot be nil")
	}

	v := reflect.ValueOf(*b)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			return fmt.Errorf("CmdHandlerParams param %v is required", typeOfS.Field(i).Name)
		}
	}

	return nil
}
