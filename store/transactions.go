package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
)

type TransactionSeq interface {
	CreateTransactionSeq(*model.TransactionSeq) error
	FindAllByHeightAndIndex(height int64, indexes []int64) (map[int64]*model.TransactionSeq, error)
	SaveTransactionSeq(*model.TransactionSeq) error
}
