package store

import (
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
	Address         string         `json:"address"`
	TimeBucket      types.Time     `json:"time_bucket"`
	VotingPowerAvg  float64        `json:"voting_power_avg"`
	VotingPowerMax  float64        `json:"voting_power_max"`
	VotingPowerMin  float64        `json:"voting_power_min"`
	TotalSharesAvg  types.Quantity `json:"total_shares_avg"`
	TotalSharesMax  types.Quantity `json:"total_shares_max"`
	TotalSharesMin  types.Quantity `json:"total_shares_min"`
	ValidatedSum    int64          `json:"validated_sum"`
	NotValidatedSum int64          `json:"not_validated_sum"`
	ProposedSum     int64          `json:"proposed_sum"`
	UptimeAvg       float64        `json:"uptime_avg"`
}

// Summarize gets the summarized version of validator sequences
//func (s *ValidatorEraSeqStore) Summarize(interval types.SummaryInterval, activityPeriods []ActivityPeriodRow) ([]ValidatorEraSeqSummary, error) {
//	defer logQueryDuration(time.Now(), "ValidatorEraSeqStore_Summarize")
//
//	tx := s.db.
//		Table(model.ValidatorEraSeq{}.TableName()).
//		Select(summarizeValidatorsQuerySelect, interval).
//		Order("time_bucket").
//		Group("address, time_bucket")
//
//	if len(activityPeriods) == 1 {
//		activityPeriod := activityPeriods[0]
//		tx = tx.Or("time < ? OR time >= ?", activityPeriod.Min, activityPeriod.Max)
//	} else {
//		for i, activityPeriod := range activityPeriods {
//			isLast := i == len(activityPeriods)-1
//
//			if isLast {
//				tx = tx.Or("time >= ?", activityPeriod.Max)
//			} else {
//				duration, err := time.ParseDuration(fmt.Sprintf("1%s", interval))
//				if err != nil {
//					return nil, err
//				}
//				tx = tx.Or("time >= ? AND time < ?", activityPeriod.Max.Add(duration), activityPeriods[i+1].Min)
//			}
//		}
//	}
//
//	rows, err := tx.Rows()
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var models []ValidatorEraSeqSummary
//	for rows.Next() {
//		var summary ValidatorEraSeqSummary
//		if err := s.db.ScanRows(rows, &summary); err != nil {
//			return nil, err
//		}
//
//		models = append(models, summary)
//	}
//	return models, nil
//}
