package store

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/jinzhu/gorm"
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewValidatorEraSeqStore(db *gorm.DB) *ValidatorEraSeqStore {
	return &ValidatorEraSeqStore{scoped(db, model.ValidatorEraSeq{})}
}

// ValidatorEraSeqStore handles operations on validators
type ValidatorEraSeqStore struct {
	baseStore
}

// CreateIfNotExists creates the validator if it does not exist
func (s ValidatorEraSeqStore) CreateIfNotExists(validator *model.ValidatorEraSeq) error {
	_, err := s.FindByHeight(validator.Era)
	if isNotFound(err) {
		return s.Create(validator)
	}
	return nil
}

// FindByHeightAndStashAccount finds validator by height and stash account
func (s ValidatorEraSeqStore) FindByHeightAndStashAccount(height int64, stash string) (*model.ValidatorEraSeq, error) {
	q := model.ValidatorEraSeq{
		StashAccount: stash,
	}
	var result model.ValidatorEraSeq

	err := s.db.
		Where(&q).
		Where("start_height >= ? AND end_height <= ?", height, height).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindByEraAndStashAccount finds validator by session and stash account
func (s ValidatorEraSeqStore) FindByEraAndStashAccount(session int64, stash string) (*model.ValidatorEraSeq, error) {
	q := model.ValidatorEraSeq{
		EraSequence: &model.EraSequence{
			Era: session,
		},
		StashAccount: stash,
	}
	var result model.ValidatorEraSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindByHeight finds validator session sequences by height
func (s ValidatorEraSeqStore) FindByHeight(h int64) ([]model.ValidatorEraSeq, error) {
	var result []model.ValidatorEraSeq

	err := s.db.
		Where("start_height >= ? AND end_height <= ?", h, h).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindByHeight finds validator session sequences by session
func (s ValidatorEraSeqStore) FindByEra(session int64) ([]model.ValidatorEraSeq, error) {
	q := model.ValidatorEraSeq{
		EraSequence: &model.EraSequence{
			Era: session,
		},
	}
	var result []model.ValidatorEraSeq

	err := s.db.
		Where(&q).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindMostRecent finds most recent validator session sequence
func (s *ValidatorEraSeqStore) FindMostRecent() (*model.ValidatorEraSeq, error) {
	validatorSeq := &model.ValidatorEraSeq{}
	if err := findMostRecent(s.db, "time", validatorSeq); err != nil {
		return nil, err
	}
	return validatorSeq, nil
}

// FindLastByStashAccount finds last validator session sequences for given stash account
func (s ValidatorEraSeqStore) FindLastByStashAccount(stashAccount string, limit int64) ([]model.ValidatorEraSeq, error) {
	q := model.ValidatorEraSeq{
		StashAccount: stashAccount,
	}
	var result []model.ValidatorEraSeq

	err := s.db.
		Where(&q).
		Order("height DESC").
		Limit(limit).
		Find(&result).
		Error

	return result, checkErr(err)
}

// DeleteOlderThan deletes validator sequence older than given threshold
func (s *ValidatorEraSeqStore) DeleteOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.ValidatorEraSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}

type ValidatorEraSeqSummary struct {
	StashAccount    string         `json:"stash_account"`
	TimeBucket      types.Time     `json:"time_bucket"`
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
}

// Summarize gets the summarized version of validator sequences
func (s *ValidatorEraSeqStore) Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]ValidatorEraSeqSummary, error) {
	defer logQueryDuration(time.Now(), "ValidatorEraSeqStore_Summarize")

	tx := s.db.
		Table(model.ValidatorEraSeq{}.TableName()).
		Select(summarizeValidatorsForEraQuerySelect, interval).
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
				duration, err := time.ParseDuration(fmt.Sprintf("1%s", interval))
				if err != nil {
					return nil, err
				}
				tx = tx.Or("time >= ? AND time < ?", activityPeriod.Max.Add(duration), activityPeriods[i+1].Min)
			}
		}
	}

	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []ValidatorEraSeqSummary
	for rows.Next() {
		var summary ValidatorEraSeqSummary
		if err := s.db.ScanRows(rows, &summary); err != nil {
			return nil, err
		}

		models = append(models, summary)
	}
	return models, nil
}
