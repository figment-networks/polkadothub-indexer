package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
)

type AccountEraSeq interface {
	baseStore
	DeleteOlderThan(purgeThreshold time.Time) (*int64, error)
	FindByEra(era int64) ([]model.AccountEraSeq, error)
	FindByEraAndStashAccount(era int64, stash string) (*model.AccountEraSeq, error)
	FindByEraAndStashAccounts(era int64, stash string, validatorStash string) (*model.AccountEraSeq, error)
	FindByHeight(h int64) ([]model.AccountEraSeq, error)
	FindByHeightAndStashAccounts(height int64, stash string, validatorStash string) (*model.AccountEraSeq, error)
	FindLastByStashAccount(stashAccount string) ([]model.AccountEraSeq, error)
	FindLastByValidatorStashAccount(validatorStashAccount string) ([]model.AccountEraSeq, error)
	FindMostRecent() (*model.AccountEraSeq, error)
}
