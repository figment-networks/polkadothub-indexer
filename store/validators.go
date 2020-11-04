package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type ValidatorAgg interface {
	CreateAgg(*model.ValidatorAgg) error
	FindBy(key string, value interface{}) (*model.ValidatorAgg, error)
	FindByID(id int64) (*model.ValidatorAgg, error)
	FindAggByStashAccount(key string) (*model.ValidatorAgg, error)
	GetAllForHeightGreaterThan(height int64) ([]model.ValidatorAgg, error)
	SaveAgg(*model.ValidatorAgg) error
}

type ValidatorSeq interface {
	CreateSeq(*model.ValidatorSeq) error
	DeleteSeqsOlderThan(purgeThreshold time.Time) (*int64, error)
	FindAllByHeight(height int64) ([]model.ValidatorSeq, error)
	FindMostRecentSeq() (*model.ValidatorSeq, error)
	SaveSeq(*model.ValidatorSeq) error
}

type ValidatorEraSeq interface {
	CreateEraSeq(*model.ValidatorEraSeq) error
	DeleteEraSeqsOlderThan(purgeThreshold time.Time) (*int64, error)
	FindByEraAndStashAccount(era int64, stash string) (*model.ValidatorEraSeq, error)
	FindEraSeqsByHeight(h int64) ([]model.ValidatorEraSeq, error)
	FindByEra(era int64) ([]model.ValidatorEraSeq, error)
	FindMostRecentEraSeq() (*model.ValidatorEraSeq, error)
	FindLastEraSeqByStashAccount(stashAccount string, limit int64) ([]model.ValidatorEraSeq, error)
	SaveEraSeq(*model.ValidatorEraSeq) error
	SummarizeEraSeqs(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]model.ValidatorEraSeqSummary, error)
}

type ValidatorSessionSeq interface {
	CreateSessionSeq(*model.ValidatorSessionSeq) error
	DeleteSessionSeqsOlderThan(purgeThreshold time.Time) (*int64, error)
	FindSessionSeqsByHeight(h int64) ([]model.ValidatorSessionSeq, error)
	FindBySession(h int64) ([]model.ValidatorSessionSeq, error)
	FindBySessionAndStashAccount(session int64, stash string) (*model.ValidatorSessionSeq, error)
	FindLastSessionSeqByStashAccount(stashAccount string, limit int64) ([]model.ValidatorSessionSeq, error)
	FindMostRecentSessionSeq() (*model.ValidatorSessionSeq, error)
	GetCountsForAccounts(sinceSession int64) (map[string]int64, error)
	SaveSessionSeq(*model.ValidatorSessionSeq) error
	SummarizeSessionSeqs(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]model.ValidatorSessionSeqSummary, error)
}

type ValidatorSummary interface {
	CreateSummary(*model.ValidatorSummary) error
	DeleteSummaryOlderThan(interval types.SummaryInterval, purgeThreshold time.Time) (*int64, error)
	FindSummary(query *model.ValidatorSummary) (*model.ValidatorSummary, error)
	FindActivityPeriods(interval types.SummaryInterval, indexVersion int64) ([]ActivityPeriodRow, error)
	FindMostRecentSummary() (*model.ValidatorSummary, error)
	FindMostRecentByInterval(interval types.SummaryInterval) (*model.ValidatorSummary, error)
	FindSummaries(interval types.SummaryInterval, period string) ([]ValidatorSummaryRow, error)
	FindSummaryByStashAccount(stashAccount string, interval types.SummaryInterval, period string) ([]ValidatorSummaryRow, error)
	SaveSummary(*model.ValidatorSummary) error
}

type ValidatorSummaryRow struct {
	TimeBucket      string         `json:"time_bucket"`
	TimeInterval    string         `json:"time_interval"`
	TotalStakeAvg   types.Quantity `json:"total_stake_avg"`
	TotalStakeMin   types.Quantity `json:"total_stake_min"`
	TotalStakeMax   types.Quantity `json:"total_stake_max"`
	OwnStakeAvg     types.Quantity `json:"own_stake_avg"`
	OwnStakeMin     types.Quantity `json:"own_stake_min"`
	OwnStakeMax     types.Quantity `json:"own_stake_max"`
	StakersStakeAvg types.Quantity `json:"stakers_stake_avg"`
	StakersStakeMin types.Quantity `json:"stakers_stake_min"`
	StakersStakeMax types.Quantity `json:"stakers_stake_max"`
	RewardPointsAvg float64        `json:"reward_points_avg"`
	RewardPointsMin int64          `json:"reward_points_min"`
	RewardPointsMax int64          `json:"reward_points_max"`
	CommissionAvg   float64        `json:"commission_avg"`
	CommissionMin   int64          `json:"commission_min"`
	CommissionMax   int64          `json:"commission_max"`
	StakersCountAvg float64        `json:"stakers_count_avg"`
	StakersCountMin int64          `json:"stakers_count_min"`
	StakersCountMax int64          `json:"stakers_count_max"`
	UptimeAvg       float64        `json:"uptime_avg"`
}
