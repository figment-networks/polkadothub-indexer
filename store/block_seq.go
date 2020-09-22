package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type BlockSeq interface {
	BaseStore
	DeleteOlderThan(purgeThreshold time.Time, activityPeriods []ActivityPeriodRow) (*int64, error)
	FindByHeight(height int64) (*model.BlockSeq, error)
	FindByID(id int64) (*model.BlockSeq, error)
	FindMostRecent() (*model.BlockSeq, error)
	GetAvgRecentTimes(limit int64) GetAvgRecentTimesResult
	Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]BlockSeqSummary, error)
}

type BlockSeqSummary struct {
	TimeBucket   types.Time `json:"time_bucket"`
	Count        int64      `json:"count"`
	BlockTimeAvg float64    `json:"block_time_avg"`
}

// GetAvgRecentTimesResult Contains results for GetAvgRecentTimes query
type GetAvgRecentTimesResult struct {
	StartHeight int64   `json:"start_height"`
	EndHeight   int64   `json:"end_height"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Count       int64   `json:"count"`
	Diff        float64 `json:"diff"`
	Avg         float64 `json:"avg"`
}
