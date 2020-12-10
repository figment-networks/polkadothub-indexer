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
	"github.com/figment-networks/polkadothub-indexer/usecase/reward"
	"github.com/figment-networks/polkadothub-indexer/usecase/system_event"
	"github.com/figment-networks/polkadothub-indexer/usecase/transaction"
	"github.com/figment-networks/polkadothub-indexer/usecase/validator"
)

func NewHttpHandlers(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *HttpHandlers {
	return &HttpHandlers{
		Health:                     health.NewHealthHttpHandler(),
		GetStatus:                  chain.NewGetStatusHttpHandler(cli, syncableDb),
		GetBlockByHeight:           block.NewGetByHeightHttpHandler(cli, syncableDb),
		GetBlockTimes:              block.NewGetBlockTimesHttpHandler(blockDb),
		GetBlockSummary:            block.NewGetBlockSummaryHttpHandler(blockDb),
		GetTransactionsByHeight:    transaction.NewGetByHeightHttpHandler(cli, syncableDb),
		GetAccountByHeight:         account.NewGetByHeightHttpHandler(cli, syncableDb),
		GetAccountDetails:          account.NewGetDetailsHttpHandler(cli, accountDb, eventDb, syncableDb),
		GetSystemEventsForAddress:  system_event.NewGetForAddressHttpHandler(cli, systemEventDb),
		GetValidatorsByHeight:      validator.NewGetByHeightHttpHandler(cfg, cli, accountDb, blockDb, databaseDb, eventDb, reportDb, rewardDb, syncableDb, systemEventDb, transactionDb, validatorDb),
		GetValidatorByStashAccount: validator.NewGetByStashAccountHttpHandler(accountDb, validatorDb),
		GetValidatorSummary:        validator.NewGetSummaryHttpHandler(syncableDb, validatorDb),
		GetValidatorsForMinHeight:  validator.NewGetForMinHeightHttpHandler(syncableDb, validatorDb),
		GetRewardsForStashAccount:  reward.NewGetForStashAccountHttpHandler(rewardDb),
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
	GetSystemEventsForAddress  types.HttpHandler
	GetValidatorsByHeight      types.HttpHandler
	GetValidatorByStashAccount types.HttpHandler
	GetValidatorSummary        types.HttpHandler
	GetValidatorsForMinHeight  types.HttpHandler
	GetRewardsForStashAccount  types.HttpHandler
}
