package usecase

import (
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

func NewHttpHandlers(cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, eventDb store.Events, syncableDb store.Syncables, validatorDb store.Validators) (*HttpHandlers, error) {
	return &HttpHandlers{
		Health:                     health.NewHealthHttpHandler(),
		GetStatus:                  chain.NewGetStatusHttpHandler(cli, syncableDb),
		GetBlockByHeight:           block.NewGetByHeightHttpHandler(cli, syncableDb),
		GetBlockTimes:              block.NewGetBlockTimesHttpHandler(blockDb),
		GetBlockSummary:            block.NewGetBlockSummaryHttpHandler(blockDb),
		GetTransactionsByHeight:    transaction.NewGetByHeightHttpHandler(cli, syncableDb),
		GetAccountByHeight:         account.NewGetByHeightHttpHandler(cli, syncableDb),
		GetAccountDetails:          account.NewGetDetailsHttpHandler(cli, accountDb, eventDb),
		GetValidatorsByHeight:      validator.NewGetByHeightHttpHandler(syncableDb, validatorDb),
		GetValidatorByStashAccount: validator.NewGetByStashAccountHttpHandler(accountDb, validatorDb),
		GetValidatorSummary:        validator.NewGetSummaryHttpHandler(validatorDb),
		GetValidatorsForMinHeight:  validator.NewGetForMinHeightHttpHandler(syncableDb, validatorDb),
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
