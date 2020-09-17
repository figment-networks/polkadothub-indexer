package psql

import (
	"github.com/figment-networks/polkadothub-indexer/store"
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

// FindSummary Gets average block times for interval
func (s *DatabaseStore) GetTotalSize() (*store.GetTotalSizeResult, error) {
	query := "SELECT pg_database_size(current_database()) as size"

	var result store.GetTotalSizeResult
	err := s.db.Raw(query).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
