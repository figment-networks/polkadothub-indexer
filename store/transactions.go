package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
)

type TransactionSeq interface {
	BulkUpsert(records []model.TransactionSeq) error
	GetTransactionByTransactionKind(kind model.TransactionKind) ([]model.TransactionSeq, error)
}
