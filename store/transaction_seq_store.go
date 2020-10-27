package store

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

// FindByHeightAndIndex finds event by height and index
func (s TransactionSeqStore) FindByHeightAndIndex(height int64, index int64) (*model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Sequence: &model.Sequence{
			Height: height,
		},
		Index: index,
	}
	var result model.TransactionSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}
