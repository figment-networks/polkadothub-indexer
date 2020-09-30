package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type ValidatorSessionSeq interface {
	BaseStore
	DeleteOlderThan(purgeThreshold time.Time) (*int64, error)
	FindByHeight(h int64) ([]model.ValidatorSessionSeq, error)
	FindBySessionAndStashAccount(session int64, stash string) (*model.ValidatorSessionSeq, error)
	FindLastByStashAccount(stashAccount string, limit int64) ([]model.ValidatorSessionSeq, error)
	FindMostRecent() (*model.ValidatorSessionSeq, error)
	Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]model.ValidatorSessionSeqSummary, error)
}
