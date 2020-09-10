package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
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

// CreateIfNotExists creates the validator if it does not exist
func (s ValidatorSessionSeqStore) CreateIfNotExists(validator *model.ValidatorSessionSeq) error {
	_, err := s.FindByHeight(validator.Session)
	if isNotFound(err) {
		return s.Create(validator)
	}
	return nil
}

// FindByHeightAndStashAccount finds validator by height and stash account
func (s ValidatorSessionSeqStore) FindByHeightAndStashAccount(height int64, stash string) (*model.ValidatorSessionSeq, error) {
	q := model.ValidatorSessionSeq{
		StashAccount: stash,
	}
	var result model.ValidatorSessionSeq

	err := s.db.
		Where(&q).
		Where("start_height >= ? AND end_height <= ?", height, height).
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

// FindByHeight finds validator session sequences by height
func (s ValidatorSessionSeqStore) FindByHeight(h int64) ([]model.ValidatorSessionSeq, error) {
	var result []model.ValidatorSessionSeq

	err := s.db.
		Where("start_height >= ? AND end_height <= ?", h, h).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindByHeight finds validator session sequences by session
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

// FindLastByStashAccount finds last validator session sequences for given stash account
func (s ValidatorSessionSeqStore) FindLastByStashAccount(stashAccount string, limit int64) ([]model.ValidatorSessionSeq, error) {
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

// FindMostRecent finds most recent validator session sequence
func (s *ValidatorSessionSeqStore) FindMostRecent() (*model.ValidatorSessionSeq, error) {
	validatorSeq := &model.ValidatorSessionSeq{}
	if err := findMostRecent(s.db, "time", validatorSeq); err != nil {
		return nil, err
	}
	return validatorSeq, nil
}

// DeleteOlderThan deletes validator sequence older than given threshold
func (s *ValidatorSessionSeqStore) DeleteOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.ValidatorSessionSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}

type ValidatorSessionSeqSummary struct {
	StashAccount string     `json:"stash_account"`
	TimeBucket   types.Time `json:"time_bucket"`
	UptimeAvg    float64    `json:"uptime_avg"`
	UptimeMin    int64      `json:"uptime_min"`
	UptimeMax    int64      `json:"uptime_max"`
}

// Summarize gets the summarized version of validator sequences
func (s *ValidatorSessionSeqStore) Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]ValidatorSessionSeqSummary, error) {
	defer logQueryDuration(time.Now(), "ValidatorSessionSeqStore_Summarize")

	tx := s.db.
		Table(model.ValidatorSessionSeq{}.TableName()).
		Select(summarizeValidatorsForSessionQuerySelect, interval).
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

	var models []ValidatorSessionSeqSummary
	for rows.Next() {
		var summary ValidatorSessionSeqSummary
		if err := s.db.ScanRows(rows, &summary); err != nil {
			return nil, err
		}

		models = append(models, summary)
	}
	return models, nil
}
