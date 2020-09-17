package psql

import (
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/jinzhu/gorm"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewBlockSeqStore(db *gorm.DB) *BlockSeqStore {
	return &BlockSeqStore{scoped(db, model.BlockSeq{})}
}

// BlockSeqStore handles operations on blocks
type BlockSeqStore struct {
	baseStore
}

// CreateIfNotExists creates the block if it does not exist
func (s BlockSeqStore) CreateIfNotExists(block *model.BlockSeq) error {
	_, err := s.FindByHeight(block.Height)
	if isNotFound(err) {
		return s.Create(block)
	}
	return nil
}

// FindBy returns a block for a matching attribute
func (s BlockSeqStore) FindBy(key string, value interface{}) (*model.BlockSeq, error) {
	result := &model.BlockSeq{}
	err := findBy(s.db, result, key, value)
	return result, checkErr(err)
}

// FindByID returns a block with matching ID
func (s BlockSeqStore) FindByID(id int64) (*model.BlockSeq, error) {
	return s.FindBy("id", id)
}

// FindByHeight returns a block with the matching height
func (s BlockSeqStore) FindByHeight(height int64) (*model.BlockSeq, error) {
	return s.FindBy("height", height)
}

// GetAvgRecentTimes Gets average block times for recent blocks by limit
func (s *BlockSeqStore) GetAvgRecentTimes(limit int64) store.GetAvgRecentTimesResult {
	defer logQueryDuration(time.Now(), "BlockSeqStore_GetAvgRecentTimes")

	var res store.GetAvgRecentTimesResult
	s.db.Raw(blockTimesForRecentBlocksQuery, limit).Scan(&res)

	return res
}

// GetAvgTimesForIntervalRow Contains row of data for FindSummary query
type GetAvgTimesForIntervalRow struct {
	TimeInterval string  `json:"time_interval"`
	Count        int64   `json:"count"`
	Avg          float64 `json:"avg"`
}

// FindMostRecent finds most recent block sequence
func (s *BlockSeqStore) FindMostRecent() (*model.BlockSeq, error) {
	blockSeq := &model.BlockSeq{}
	if err := findMostRecent(s.db, "time", blockSeq); err != nil {
		return nil, err
	}
	return blockSeq, nil
}

// DeleteOlderThan deletes block sequence older than given threshold
func (s *BlockSeqStore) DeleteOlderThan(purgeThreshold time.Time, activityPeriods []store.ActivityPeriodRow) (*int64, error) {
	tx := s.db.
		Unscoped()

	hasIntervals := false
	for _, activityPeriod := range activityPeriods {
		// Make sure that there are many intervals (ie. days) in period
		if !activityPeriod.Min.Equal(activityPeriod.Max) {
			hasIntervals = true
			// Thus, we do not add 1 day to Max because we don't want to purge sequences within last day of period
			tx = tx.Where("time >= ? AND time < ?", activityPeriod.Min, activityPeriod.Max)
		}
	}

	if hasIntervals {
		tx.Where("time < ?", purgeThreshold).
			Delete(&model.BlockSeq{})

		if tx.Error != nil {
			return nil, checkErr(tx.Error)
		}
	} else {
		logger.Info("no block sequences to purge")
	}

	return &tx.RowsAffected, nil
}

// Summarize gets the summarized version of block sequences
func (s *BlockSeqStore) Summarize(interval types.SummaryInterval, activityPeriods []store.ActivityPeriodRow) ([]store.BlockSeqSummary, error) {
	defer logQueryDuration(time.Now(), "BlockSummaryStore_Summarize")

	tx := s.db.
		Table(model.BlockSeq{}.TableName()).
		Select(summarizeBlocksQuerySelect, interval).
		Order("time_bucket").
		Group("time_bucket")

	fmt.Println("INT", interval, activityPeriods)

	if len(activityPeriods) == 1 {
		activityPeriod := activityPeriods[0]
		tx = tx.Or("time < ? OR time >= ?", activityPeriod.Min, activityPeriod.Max)
	} else {
		for i, activityPeriod := range activityPeriods {
			isLast := i == len(activityPeriods)-1

			if isLast {
				tx = tx.Or("time >= ?", activityPeriod.Max)
			} else {
				duration, err := interval.Duration()
				if err != nil {
					return nil, err
				}
				tx = tx.Or("time >= ? AND time < ?", activityPeriod.Max.Add(*duration), activityPeriods[i+1].Min)
			}
		}
	}

	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []store.BlockSeqSummary
	for rows.Next() {
		var summary store.BlockSeqSummary
		if err := s.db.ScanRows(rows, &summary); err != nil {
			return nil, err
		}

		models = append(models, summary)
	}
	return models, nil
}
