package usecase

import (
	"fmt"
	"reflect"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/indexing"
)

type WorkerHandlerParams struct {
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

func NewWorkerHandlers(b *WorkerHandlerParams) (*WorkerHandlers, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}

	return &WorkerHandlers{
		RunIndexer: indexing.NewRunWorkerHandler(b.Config, b.Client, b.AccountEraSeqDb, b.BlockSeqDb, b.BlockSummaryDb, b.DatabaseDb, b.EventSeqDb, b.ReportsDb,
			b.SyncablesDb, b.ValidatorAggDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb,
		),
		SummarizeIndexer: indexing.NewSummarizeWorkerHandler(b.Config, b.BlockSeqDb, b.BlockSummaryDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb),
		PurgeIndexer:     indexing.NewPurgeWorkerHandler(b.Config, b.BlockSeqDb, b.BlockSummaryDb, b.ValidatorSessionSeqDb, b.ValidatorSummaryDb),
	}, nil
}

type WorkerHandlers struct {
	RunIndexer       types.WorkerHandler
	SummarizeIndexer types.WorkerHandler
	PurgeIndexer     types.WorkerHandler
}

func (b *WorkerHandlerParams) validate() error {
	if b == nil {
		return fmt.Errorf("WorkerHandlerParams cannot be nil")
	}

	v := reflect.ValueOf(*b)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			return fmt.Errorf("WorkerHandlerParams param %v is required", typeOfS.Field(i).Name)
		}
	}

	return nil
}
