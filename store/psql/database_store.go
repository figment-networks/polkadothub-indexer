package psql

import (
	"github.com/jinzhu/gorm"
)

func NewDatabaseStore(db *gorm.DB) *DatabaseStore {
	return &DatabaseStore{
		db: db,
	}
}

// DatabaseStore handles operations on blocks
type DatabaseStore struct {
	db *gorm.DB
}

// GetAvgTimesForIntervalRow Contains row of data for FindSummary query
type GetTotalSizeResult struct {
	Size float64 `json:"size"`
}

// FindSummary Gets average block times for interval
func (s *DatabaseStore) GetTotalSize() (*GetTotalSizeResult, error) {
	query := "SELECT pg_database_size(current_database()) as size"

	var result GetTotalSizeResult
	err := s.db.Raw(query).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
