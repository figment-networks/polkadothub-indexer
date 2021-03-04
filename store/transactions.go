package store

import (
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/model"
)

type TransactionSeq interface {
	BulkUpsert(records []model.TransactionSeq) error
	GetTransactionByTrxFilter(filter indexer.TrxFilter) ([]model.TransactionSeq, error)
}
