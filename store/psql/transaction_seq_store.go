package psql

import (
	"github.com/jinzhu/gorm"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewTransactionSeqStore(db *gorm.DB) *TransactionSeqStore {
	return &TransactionSeqStore{scoped(db, model.TransactionSeq{})}
}

// TransactionSeqStore handles operations on balance events
type TransactionSeqStore struct {
	baseStore
}

// CreateTransactionSeq creates the validator aggregate
func (s TransactionSeqStore) CreateTransactionSeq(val *model.TransactionSeq) error {
	return s.Create(val)
}

// SaveTransactionSeq creates the validator aggregate
func (s TransactionSeqStore) SaveTransactionSeq(val *model.TransactionSeq) error {
	return s.Save(val)
}

// FindAllByHeightAndIndex finds all found sequences for indexes at given height, it returns map with all found sequences with indexes as keys
func (s TransactionSeqStore) FindAllByHeightAndIndex(height int64, indexes []int64) (map[int64]*model.TransactionSeq, error) {
	var results []*model.TransactionSeq

	query := getFindAllByHeightAndIndexQuery(model.TransactionSeq{}.TableName())
	err := s.db.
		Raw(query, height, indexes).
		Find(&results).
		Error

	resultMap := make(map[int64]*model.TransactionSeq, len(results))
	for _, result := range results {
		resultMap[result.Index] = result
	}

	return resultMap, checkErr(err)
}
