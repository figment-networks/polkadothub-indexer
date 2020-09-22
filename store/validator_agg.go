package store

import "github.com/figment-networks/polkadothub-indexer/model"

type ValidatorAgg interface {
	BaseStore
	All() ([]model.ValidatorAgg, error)
	CreateOrUpdate(val *model.ValidatorAgg) error
	FindBy(key string, value interface{}) (*model.ValidatorAgg, error)
	FindByID(id int64) (*model.ValidatorAgg, error)
	FindByStashAccount(key string) (*model.ValidatorAgg, error)
	GetAllForHeightGreaterThan(height int64) ([]model.ValidatorAgg, error)
}
