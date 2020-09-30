package psql

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/jinzhu/gorm"
)

func NewValidatorAggStore(db *gorm.DB) *ValidatorAggStore {
	return &ValidatorAggStore{scoped(db, model.ValidatorAgg{})}
}

// ValidatorAgg handles operations on validators
type ValidatorAggStore struct {
	baseStore
}

// CreateAgg creates the validator aggregate
func (s ValidatorAggStore) CreateAgg(val *model.ValidatorAgg) error {
	return s.Create(val)
}

// SaveAgg creates the validator aggregate
func (s ValidatorAggStore) SaveAgg(val *model.ValidatorAgg) error {
	return s.Save(val)
}

// FindBy returns an validator for a matching attribute
func (s ValidatorAggStore) FindBy(key string, value interface{}) (*model.ValidatorAgg, error) {
	result := &model.ValidatorAgg{}
	err := findBy(s.db, result, key, value)
	return result, checkErr(err)
}

// FindByID returns an validator for the ID
func (s ValidatorAggStore) FindByID(id int64) (*model.ValidatorAgg, error) {
	return s.FindBy("id", id)
}

// FindAggByStashAccount return validator by stash account
func (s *ValidatorAggStore) FindAggByStashAccount(key string) (*model.ValidatorAgg, error) {
	return s.FindBy("stash_account", key)
}

// GetAllForHeightGreaterThan returns validators who have been validating since given height
func (s *ValidatorAggStore) GetAllForHeightGreaterThan(height int64) ([]model.ValidatorAgg, error) {
	var result []model.ValidatorAgg

	err := s.baseStore.db.
		Where("recent_as_validator_height >= ?", height).
		Find(&result).
		Error

	return result, checkErr(err)
}

// All returns all validators
func (s ValidatorAggStore) All() ([]model.ValidatorAgg, error) {
	var result []model.ValidatorAgg

	err := s.db.
		Order("id ASC").
		Find(&result).
		Error

	return result, checkErr(err)
}
