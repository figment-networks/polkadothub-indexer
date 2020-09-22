package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type BlockSummary interface {
	BaseStore
	DeleteOlderThan(interval types.SummaryInterval, purgeThreshold time.Time) (*int64, error)
	Find(query *model.BlockSummary) (*model.BlockSummary, error)
	FindActivityPeriods(interval types.SummaryInterval, indexVersion int64) ([]ActivityPeriodRow, error)
	FindMostRecent() (*model.BlockSummary, error)
	FindMostRecentByInterval(interval types.SummaryInterval) (*model.BlockSummary, error)
	FindSummary(interval types.SummaryInterval, period string) ([]model.BlockSummary, error)
}
