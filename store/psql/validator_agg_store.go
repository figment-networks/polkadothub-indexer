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

// CreateOrUpdate creates a new validator or updates an existing one
func (s ValidatorAggStore) CreateOrUpdate(val *model.ValidatorAgg) error {
	existing, err := s.FindByStashAccount(val.StashAccount)
	if err != nil {
		if err == ErrNotFound {
			return s.Create(val)
		}
		return err
	}
	return s.Update(existing)
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

// FindByStashAccount return validator by stash account
func (s *ValidatorAggStore) FindByStashAccount(key string) (*model.ValidatorAgg, error) {
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
