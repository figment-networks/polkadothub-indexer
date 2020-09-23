package usecase

import (
	"fmt"
	"reflect"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/account"
	"github.com/figment-networks/polkadothub-indexer/usecase/block"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/health"
	"github.com/figment-networks/polkadothub-indexer/usecase/transaction"
	"github.com/figment-networks/polkadothub-indexer/usecase/validator"
)

type HttpHandlerParams struct {
	Client                *client.Client
	AccountEraSeqDb       store.AccountEraSeq
	BlockSeqDb            store.BlockSeq
	BlockSummaryDb        store.BlockSummary
	EventSeqDb            store.EventSeq
	SyncablesDb           store.Syncables
	ValidatorAggDb        store.ValidatorAgg
	ValidatorEraSeqDb     store.ValidatorEraSeq
	ValidatorSessionSeqDb store.ValidatorSessionSeq
	ValidatorSummaryDb    store.ValidatorSummary
}

func NewHttpHandlers(b *HttpHandlerParams) (*HttpHandlers, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}

	return &HttpHandlers{
		Health:                     health.NewHealthHttpHandler(),
		GetStatus:                  chain.NewGetStatusHttpHandler(b.Client, b.SyncablesDb),
		GetBlockByHeight:           block.NewGetByHeightHttpHandler(b.Client, b.SyncablesDb),
		GetBlockTimes:              block.NewGetBlockTimesHttpHandler(b.BlockSeqDb),
		GetBlockSummary:            block.NewGetBlockSummaryHttpHandler(b.BlockSummaryDb),
		GetTransactionsByHeight:    transaction.NewGetByHeightHttpHandler(b.Client, b.SyncablesDb),
		GetAccountByHeight:         account.NewGetByHeightHttpHandler(b.Client, b.SyncablesDb),
		GetAccountDetails:          account.NewGetDetailsHttpHandler(b.Client, b.AccountEraSeqDb, b.EventSeqDb),
		GetValidatorsByHeight:      validator.NewGetByHeightHttpHandler(b.SyncablesDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb),
		GetValidatorByStashAccount: validator.NewGetByStashAccountHttpHandler(b.AccountEraSeqDb, b.ValidatorAggDb, b.ValidatorEraSeqDb, b.ValidatorSessionSeqDb),
		GetValidatorSummary:        validator.NewGetSummaryHttpHandler(b.ValidatorSummaryDb),
		GetValidatorsForMinHeight:  validator.NewGetForMinHeightHttpHandler(b.SyncablesDb, b.ValidatorAggDb),
	}, nil
}

type HttpHandlers struct {
	Health                     types.HttpHandler
	GetStatus                  types.HttpHandler
	GetBlockTimes              types.HttpHandler
	GetBlockSummary            types.HttpHandler
	GetBlockByHeight           types.HttpHandler
	GetTransactionsByHeight    types.HttpHandler
	GetAccountByHeight         types.HttpHandler
	GetAccountDetails          types.HttpHandler
	GetValidatorsByHeight      types.HttpHandler
	GetValidatorByStashAccount types.HttpHandler
	GetValidatorSummary        types.HttpHandler
	GetValidatorsForMinHeight  types.HttpHandler
}

func (b *HttpHandlerParams) validate() error {
	if b == nil {
		return fmt.Errorf("HttpHandlerParams cannot be nil")
	}

	v := reflect.ValueOf(*b)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			return fmt.Errorf("HttpHandlerParams param %v is required", typeOfS.Field(i).Name)
		}
	}

	return nil
}
