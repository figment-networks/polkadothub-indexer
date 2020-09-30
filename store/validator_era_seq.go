package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type ValidatorEraSeq interface {
	BaseStore
	CreateIfNotExists(validator *model.ValidatorEraSeq) error
	FindByHeightAndStashAccount(height int64, stash string) (*model.ValidatorEraSeq, error)
	FindByEraAndStashAccount(era int64, stash string) (*model.ValidatorEraSeq, error)
	FindByHeight(h int64) ([]model.ValidatorEraSeq, error)
	FindByEra(era int64) ([]model.ValidatorEraSeq, error)
	FindMostRecent() (*model.ValidatorEraSeq, error)
	FindLastByStashAccount(stashAccount string, limit int64) ([]model.ValidatorEraSeq, error)
	DeleteOlderThan(purgeThreshold time.Time) (*int64, error)
	Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]model.ValidatorEraSeqSummary, error)
}
