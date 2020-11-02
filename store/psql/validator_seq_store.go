package psql

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/jinzhu/gorm"
)

func NewValidatorSeqStore(db *gorm.DB) *ValidatorSeqStore {
	return &ValidatorSeqStore{scoped(db, model.ValidatorSeq{})}
}

// ValidatorSeqStore handles operations on validators
type ValidatorSeqStore struct {
	baseStore
}

// CreateSeq creates the validator aggregate
func (s ValidatorSeqStore) CreateSeq(val *model.ValidatorSeq) error {
	return s.Create(val)
}

// SaveSeq creates the validator aggregate
func (s ValidatorSeqStore) SaveSeq(val *model.ValidatorSeq) error {
	return s.Save(val)
}

// FindAllByHeight returns all validators for provided height
func (s ValidatorSeqStore) FindAllByHeight(height int64) ([]model.ValidatorSeq, error) {
	var results []model.ValidatorSeq

	err := s.db.
		Where("height = ?", height).
		Find(&results).
		Error

	return results, checkErr(err)
}
