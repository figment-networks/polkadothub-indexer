package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type AccountEraSeq interface {
	BulkUpsert(records []model.AccountEraSeq) error
	FindByEra(era int64) ([]model.AccountEraSeq, error)
	FindLastByStashAccount(stashAccount string) ([]model.AccountEraSeq, error)
	FindLastByValidatorStashAccount(validatorStashAccount string) ([]model.AccountEraSeq, error)
	GetAllByTime(stash string, start, end types.Time) ([]model.AccountEraSeq, error)
}
