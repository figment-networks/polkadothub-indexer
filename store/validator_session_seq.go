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
	Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]ValidatorSessionSeqSummary, error)
}

type ValidatorSessionSeqSummary struct {
	StashAccount string     `json:"stash_account"`
	TimeBucket   types.Time `json:"time_bucket"`
	UptimeAvg    float64    `json:"uptime_avg"`
	UptimeMin    int64      `json:"uptime_min"`
	UptimeMax    int64      `json:"uptime_max"`
}
