package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type BlockSummary interface {
	BaseStore

	Find(query *model.BlockSummary) (*model.BlockSummary, error)
	FindMostRecent() (*model.BlockSummary, error)
	FindMostRecentByInterval(interval types.SummaryInterval) (*model.BlockSummary, error)
	FindActivityPeriods(interval types.SummaryInterval, indexVersion int64) ([]ActivityPeriodRow, error)
	FindSummary(interval types.SummaryInterval, period string) ([]model.BlockSummary, error)
	DeleteOlderThan(interval types.SummaryInterval, purgeThreshold time.Time) (*int64, error)
}
