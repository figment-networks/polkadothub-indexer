package usecase

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/account"
	"github.com/figment-networks/polkadothub-indexer/usecase/block"
	"github.com/figment-networks/polkadothub-indexer/usecase/chain"
	"github.com/figment-networks/polkadothub-indexer/usecase/health"
	"github.com/figment-networks/polkadothub-indexer/usecase/transaction"
	"github.com/figment-networks/polkadothub-indexer/usecase/validator"
)

func NewHttpHandlers(cfg *config.Config, db store.Store, c *client.Client) *HttpHandlers {
	return &HttpHandlers{
		Health:                     health.NewHealthHttpHandler(),
		GetStatus:                  chain.NewGetStatusHttpHandler(db, c),
		GetBlockByHeight:           block.NewGetByHeightHttpHandler(db, c),
		GetBlockTimes:              block.NewGetBlockTimesHttpHandler(db, c),
		GetBlockSummary:            block.NewGetBlockSummaryHttpHandler(db, c),
		GetTransactionsByHeight:    transaction.NewGetByHeightHttpHandler(db, c),
		GetAccountByHeight:         account.NewGetByHeightHttpHandler(db, c),
		GetAccountDetails:          account.NewGetDetailsHttpHandler(db, c),
		GetValidatorsByHeight:      validator.NewGetByHeightHttpHandler(cfg, db, c),
		GetValidatorByStashAccount: validator.NewGetByStashAccountHttpHandler(db, c),
		GetValidatorSummary:        validator.NewGetSummaryHttpHandler(db, c),
		GetValidatorsForMinHeight:  validator.NewGetForMinHeightHttpHandler(db, c),
	}
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
