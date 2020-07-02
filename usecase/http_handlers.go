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
)

func NewHttpHandlers(cfg *config.Config, db *store.Store, c *client.Client) *HttpHandlers {
	return &HttpHandlers{
		Health:                  health.NewHealthHttpHandler(),
		GetStatus:               chain.NewGetStatusHttpHandler(db, c),
		GetBlockByHeight:        block.NewGetByHeightHttpHandler(db, c),
		GetBlockTimes:           block.NewGetBlockTimesHttpHandler(db, c),
		GetBlockSummary:         block.NewGetBlockSummaryHttpHandler(db, c),
		GetTransactionsByHeight: transaction.NewGetByHeightHttpHandler(db, c),
		GetAccountByHeight:      account.NewGetByHeightHttpHandler(db, c),
		GetAccountIdentity:      account.NewGetIdentityHttpHandler(db, c),
	}
}

type HttpHandlers struct {
	Health                  types.HttpHandler
	GetStatus               types.HttpHandler
	GetBlockTimes           types.HttpHandler
	GetBlockSummary         types.HttpHandler
	GetBlockByHeight        types.HttpHandler
	GetTransactionsByHeight types.HttpHandler
	GetAccountByHeight      types.HttpHandler
	GetAccountIdentity      types.HttpHandler
}
