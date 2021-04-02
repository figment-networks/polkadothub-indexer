package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type BlockSeq interface {
	CreateSeq(*model.BlockSeq) error
	DeleteSeqOlderThan(purgeThreshold time.Time, activityPeriods []ActivityPeriodRow) (*int64, error)
	FindSeqByHeight(height int64) (*model.BlockSeq, error)
	// FindByID(id int64) (*model.BlockSeq, error)
	FindMostRecentSeq() (*model.BlockSeq, error)
	GetAvgRecentTimes(limit int64) GetAvgRecentTimesResult
	SaveSeq(*model.BlockSeq) error
	Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]model.BlockSeqSummary, error)
}

type BlockSummary interface {
	CreateSummary(val *model.BlockSummary) error
	DeleteOlderThan(interval types.SummaryInterval, purgeThreshold time.Time) (*int64, error)
	FindSummary(query *model.BlockSummary) (*model.BlockSummary, error)
	FindActivityPeriods(interval types.SummaryInterval, indexVersion int64) ([]ActivityPeriodRow, error)
	FindMostRecentSummary() (*model.BlockSummary, error)
	FindMostRecentByInterval(interval types.SummaryInterval) (*model.BlockSummary, error)
	FindSummaries(interval types.SummaryInterval, period string) ([]model.BlockSummary, error)
	SaveSummary(val *model.BlockSummary) error
}

// swagger:response BlockTimesView
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
