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

func NewHttpHandlers(c *client.Client, accountEraSeqDb store.AccountEraSeq, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary,
	eventSeqDb store.EventSeq, syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg, validatorEraSeqDb store.ValidatorEraSeq,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *HttpHandlers {
	return &HttpHandlers{
		Health:                     health.NewHealthHttpHandler(),
		GetStatus:                  chain.NewGetStatusHttpHandler(c, syncablesDb),
		GetBlockByHeight:           block.NewGetByHeightHttpHandler(c, syncablesDb),
		GetBlockTimes:              block.NewGetBlockTimesHttpHandler(blockSeqDb),
		GetBlockSummary:            block.NewGetBlockSummaryHttpHandler(blockSummaryDb),
		GetTransactionsByHeight:    transaction.NewGetByHeightHttpHandler(c, syncablesDb),
		GetAccountByHeight:         account.NewGetByHeightHttpHandler(c, syncablesDb),
		GetAccountDetails:          account.NewGetDetailsHttpHandler(c, accountEraSeqDb, eventSeqDb),
		GetValidatorsByHeight:      validator.NewGetByHeightHttpHandler(syncablesDb, validatorEraSeqDb, validatorSessionSeqDb),
		GetValidatorByStashAccount: validator.NewGetByStashAccountHttpHandler(accountEraSeqDb, validatorAggDb, validatorEraSeqDb, validatorSessionSeqDb),
		GetValidatorSummary:        validator.NewGetSummaryHttpHandler(validatorSummaryDb),
		GetValidatorsForMinHeight:  validator.NewGetForMinHeightHttpHandler(syncablesDb, validatorAggDb),
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
