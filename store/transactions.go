package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
)

type TransactionSeq interface {
	BulkUpsert(records []model.TransactionSeq) error
	GetTransactionsByTransactionKind(kind model.TransactionKind, start, end int64) ([]model.TransactionSeq, error)
}
