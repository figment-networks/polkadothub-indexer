package psql

import (
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/jinzhu/gorm"
)

func NewValidatorSummaryStore(db *gorm.DB) *ValidatorSummaryStore {
	return &ValidatorSummaryStore{scoped(db, model.ValidatorSummary{})}
}

// ValidatorSummaryStore handles operations on validators
type ValidatorSummaryStore struct {
	baseStore
}

// CreateSummary creates the validator aggregate
func (s ValidatorSummaryStore) CreateSummary(val *model.ValidatorSummary) error {
	return s.Create(val)
}

// SaveSummary creates the validator aggregate
func (s ValidatorSummaryStore) SaveSummary(val *model.ValidatorSummary) error {
	return s.Save(val)
}

// FindSummary find validator summary by query
func (s ValidatorSummaryStore) FindSummary(query *model.ValidatorSummary) (*model.ValidatorSummary, error) {
	var result model.ValidatorSummary

	err := s.db.
		Where(query).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindActivityPeriods Finds activity periods
func (s *ValidatorSummaryStore) FindActivityPeriods(interval types.SummaryInterval, indexVersion int64) ([]store.ActivityPeriodRow, error) {
	defer logQueryDuration(time.Now(), "ValidatorSummaryStore_FindActivityPeriods")

	rows, err := s.db.
		Raw(validatorSummaryActivityPeriodsQuery, fmt.Sprintf("1%s", interval), interval, indexVersion).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []store.ActivityPeriodRow
	for rows.Next() {
		var row store.ActivityPeriodRow
		if err := s.db.ScanRows(rows, &row); err != nil {
			return nil, err
		}
		res = append(res, row)
	}
	return res, nil
}

// FindSummaries gets summary for validator summary
func (s *ValidatorSummaryStore) FindSummaries(interval types.SummaryInterval, period string) ([]store.ValidatorSummaryRow, error) {
	defer logQueryDuration(time.Now(), "ValidatorSummaryStore_FindSummary")
	var res []ValidatorSummaryRow

	err := s.db.
		Raw(allValidatorsSummaryForIntervalQuery, interval, period, interval).
		Scan(&res).Error
	return res, err
}

// FindSummaryByStashAccount gets summary for given validator
func (s *ValidatorSummaryStore) FindSummaryByStashAccount(stashAccount string, interval types.SummaryInterval, period string) ([]ValidatorSummaryRow, error) {
	defer logQueryDuration(time.Now(), "ValidatorSummaryStore_FindSummaryByStashAccount")
	var res []ValidatorSummaryRow

	err := s.db.
		Raw(validatorSummaryForIntervalQuery, interval, period, stashAccount, interval).
		Scan(&res).Error
	return res, err
}

// FindMostRecentSummary finds most recent validator summary
func (s *ValidatorSummaryStore) FindMostRecentSummary() (*model.ValidatorSummary, error) {
	validatorSummary := &model.ValidatorSummary{}
	err := findMostRecent(s.db, "time_bucket", validatorSummary)
	return validatorSummary, checkErr(err)
}

// FindMostRecentByInterval finds most recent validator summary for interval
func (s *ValidatorSummaryStore) FindMostRecentByInterval(interval types.SummaryInterval) (*model.ValidatorSummary, error) {
	query := &model.ValidatorSummary{
		Summary: &model.Summary{TimeInterval: interval},
	}
	result := model.ValidatorSummary{}

	err := s.db.
		Where(query).
		Order("time_bucket DESC").
		Take(&result).
		Error

	return &result, checkErr(err)
}

// DeleteSummaryOlderThan deleted validator summary records older than given threshold
func (s *ValidatorSummaryStore) DeleteSummaryOlderThan(interval types.SummaryInterval, purgeThreshold time.Time) (*int64, error) {
	statement := s.db.
		Unscoped().
		Where("time_interval = ? AND time_bucket < ?", interval, purgeThreshold).
		Delete(&model.ValidatorSummary{})

	if statement.Error != nil {
		return nil, checkErr(statement.Error)
	}

	return &statement.RowsAffected, nil
}
