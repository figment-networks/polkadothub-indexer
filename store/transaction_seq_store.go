package store

import (
	"github.com/jinzhu/gorm"

	"github.com/figment-networks/polkadothub-indexer/model"
)

const (
	txSeqFindAllByHeightAndIndexQuery = `
	SELECT *
	FROM transaction_sequences as t
	WHERE t.height = ? AND t.index IN (?)
`
)

func NewTransactionSeqStore(db *gorm.DB) *TransactionSeqStore {
	return &TransactionSeqStore{scoped(db, model.TransactionSeq{})}
}

// TransactionSeqStore handles operations on balance events
type TransactionSeqStore struct {
	baseStore
}

func (s TransactionSeqStore) FindAllByHeightAndIndex(height int64, indexes []int64) (map[int64]*model.TransactionSeq, error) {
	var results []*model.TransactionSeq

	err := s.db.
		Raw(txSeqFindAllByHeightAndIndexQuery, height, indexes).
		Find(&results).
		Error

	resultMap := make(map[int64]*model.TransactionSeq, len(results))
	for _, result := range results {
		resultMap[result.Index] = result
	}

	return resultMap, checkErr(err)
}
