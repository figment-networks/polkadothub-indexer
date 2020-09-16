package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/jinzhu/gorm"
)

func NewReportsStore(db *gorm.DB) *ReportsStore {
	return &ReportsStore{scoped(db, model.Report{})}
}

// ReportsStore handles operations on reports
type ReportsStore struct {
	baseStore
}

// FindNotCompletedByIndexVersion returns the report by index version and kind
func (s ReportsStore) FindNotCompletedByIndexVersion(indexVersion int64, kinds ...model.ReportKind) (*model.Report, error) {
	query := &model.Report{
		IndexVersion: indexVersion,
	}
	result := &model.Report{}

	err := s.db.
		Where(query).
		Where("kind IN(?)", kinds).
		Where("completed_at IS NULL").
		First(result).Error

	return result, checkErr(err)
}

// Last returns the last report
func (s ReportsStore) FindNotCompletedByKind(kinds ...model.ReportKind) (*model.Report, error) {
	result := &model.Report{}

	err := s.db.
		Where("kind IN(?)", kinds).
		Where("completed_at IS NULL").
		First(result).Error

	return result, checkErr(err)
}

// Last returns the last report
func (s ReportsStore) Last() (*model.Report, error) {
	result := &model.Report{}

	err := s.db.
		Order("id DESC").
		First(result).Error

	return result, checkErr(err)
}

// DeleteByKinds deletes reports with kind reindexing sequential or parallel
func (s *ReportsStore) DeleteByKinds(kinds []model.ReportKind) error {
	err := s.db.
		Unscoped().
		Where("kind IN(?)", kinds).
		Delete(&model.Report{}).
		Error

	return checkErr(err)
}
