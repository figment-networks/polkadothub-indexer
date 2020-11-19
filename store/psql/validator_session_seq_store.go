package psql

import (
	"time"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/jinzhu/gorm"
)

func NewValidatorSessionSeqStore(db *gorm.DB) *ValidatorSessionSeqStore {
	return &ValidatorSessionSeqStore{scoped(db, model.ValidatorSessionSeq{})}
}

// ValidatorSessionSeqStore handles operations on validators
type ValidatorSessionSeqStore struct {
	baseStore
}

// BulkUpsertSessionSeqs imports new records and updates existing ones
func (s ValidatorSessionSeqStore) BulkUpsertSessionSeqs(records []model.ValidatorSessionSeq) error {
	return s.Import(queries.ValidatorSessionSeqInsert, len(records), func(i int) bulk.Row {
		r := records[i]
		return bulk.Row{
			r.Session,
			r.StartHeight,
			r.EndHeight,
			r.Time,
			r.StashAccount,
			r.Online,
		}
	})
}

// FindByHeightAndStashAccount finds validator by height and stash account
func (s ValidatorSessionSeqStore) FindByHeightAndStashAccount(height int64, stash string) (*model.ValidatorSessionSeq, error) {
	q := model.ValidatorSessionSeq{
		StashAccount: stash,
	}
	var result model.ValidatorSessionSeq

	err := s.db.
		Where(&q).
		Where("start_height <= ? AND end_height >= ?", height, height).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindBySessionAndStashAccount finds validator by session and stash account
func (s ValidatorSessionSeqStore) FindBySessionAndStashAccount(session int64, stash string) (*model.ValidatorSessionSeq, error) {
	q := model.ValidatorSessionSeq{
		SessionSequence: &model.SessionSequence{
			Session: session,
		},
		StashAccount: stash,
	}
	var result model.ValidatorSessionSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindSessionSeqsByHeight finds validator session sequences by height
func (s ValidatorSessionSeqStore) FindSessionSeqsByHeight(h int64) ([]model.ValidatorSessionSeq, error) {
	var result []model.ValidatorSessionSeq

	err := s.db.
		Where("start_height <= ? AND end_height >= ?", h, h).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBySession finds validator session sequences by session
func (s ValidatorSessionSeqStore) FindBySession(session int64) ([]model.ValidatorSessionSeq, error) {
	q := model.ValidatorSessionSeq{
		SessionSequence: &model.SessionSequence{
			Session: session,
		},
	}
	var result []model.ValidatorSessionSeq

	err := s.db.
		Where(&q).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindLastSessionSeqByStashAccount finds last validator session sequences for given stash account
func (s ValidatorSessionSeqStore) FindLastSessionSeqByStashAccount(stashAccount string, limit int64) ([]model.ValidatorSessionSeq, error) {
	q := model.ValidatorSessionSeq{
		StashAccount: stashAccount,
	}
	var result []model.ValidatorSessionSeq

	err := s.db.
		Where(&q).
		Order("session DESC").
		Limit(limit).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindMostRecentSessionSeq finds most recent validator session sequence
func (s *ValidatorSessionSeqStore) FindMostRecentSessionSeq() (*model.ValidatorSessionSeq, error) {
	validatorSeq := &model.ValidatorSessionSeq{}
	if err := findMostRecent(s.db, "time", validatorSeq); err != nil {
		return nil, err
	}
	return validatorSeq, nil
}

// DeleteSessionSeqsOlderThan deletes validator sequence older than given threshold
func (s *ValidatorSessionSeqStore) DeleteSessionSeqsOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.ValidatorSessionSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}

// SummarizeSessionSeqs gets the summarized version of validator sequences
func (s *ValidatorSessionSeqStore) SummarizeSessionSeqs(interval types.SummaryInterval, activityPeriods []store.ActivityPeriodRow) ([]model.ValidatorSessionSeqSummary, error) {
	defer logQueryDuration(time.Now(), "ValidatorSessionSeqStore_Summarize")

	tx := s.db.
		Table(model.ValidatorSessionSeq{}.TableName()).
		Select(queries.ValidatorSessionSeqSummarizeSelect, interval).
		Order("time_bucket").
		Group("stash_account, time_bucket")

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

	var models []model.ValidatorSessionSeqSummary
	for rows.Next() {
		var summary model.ValidatorSessionSeqSummary
		if err := s.db.ScanRows(rows, &summary); err != nil {
			return nil, err
		}

		models = append(models, summary)
	}
	return models, nil
}
