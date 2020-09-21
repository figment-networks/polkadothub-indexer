package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/jinzhu/gorm"
)

func NewSyncablesStore(db *gorm.DB) *SyncablesStore {
	return &SyncablesStore{scoped(db, model.Report{})}
}

// SyncablesStore handles operations on syncables
type SyncablesStore struct {
	baseStore
}

// FindSmallestIndexVersion returns smallest index version
func (s SyncablesStore) FindSmallestIndexVersion() (*int64, error) {
	result := &model.Syncable{}

	err := s.db.
		Where("processed_at IS NOT NULL").
		Order("index_version").
		First(result).Error

	return &result.IndexVersion, checkErr(err)
}

// Exists returns true if a syncable exists at give height
func (s SyncablesStore) FindByHeight(height int64) (syncable *model.Syncable, err error) {
	result := &model.Syncable{}

	err = s.db.
		Where("height = ?", height).
		First(result).
		Error

	return result, checkErr(err)
}

// FindMostRecent returns the most recent syncable
func (s SyncablesStore) FindMostRecent() (*model.Syncable, error) {
	result := &model.Syncable{}

	err := s.db.
		Order("height desc").
		First(result).Error

	return result, checkErr(err)
}

// FindLastInSessionForHeight finds last_in_session syncable for given height
func (s SyncablesStore) FindLastInSessionForHeight(height int64) (syncable *model.Syncable, err error) {
	result := &model.Syncable{}

	err = s.db.
		Where("height >= ? AND last_in_session = ?", height, true).
		Order("height DESC").
		First(result).
		Error

	return result, checkErr(err)
}

// FindLastInEraForHeight finds last_in_era syncable for given height
func (s SyncablesStore) FindLastInEraForHeight(height int64) (syncable *model.Syncable, err error) {
	result := &model.Syncable{}

	err = s.db.
		Where("height >= ? AND last_in_era = ?", height, true).
		Order("height DESC").
		First(result).
		Error

	return result, checkErr(err)
}

// FindLastInSession finds last syncable in given session
func (s SyncablesStore) FindLastInSession(session int64) (syncable *model.Syncable, err error) {
	result := &model.Syncable{}

	err = s.db.
		Where("session = ?", session).
		Order("height DESC").
		First(result).
		Error

	return result, checkErr(err)
}

// FindLastInEra finds last syncable in given era
func (s SyncablesStore) FindLastInEra(era int64) (syncable *model.Syncable, err error) {
	result := &model.Syncable{}

	err = s.db.
		Where("era = ?", era).
		Order("height DESC").
		First(result).
		Error

	return result, checkErr(err)
}

// FindFirstByDifferentIndexVersion returns first syncable with different index version
func (s SyncablesStore) FindFirstByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error) {
	result := &model.Syncable{}

	err := s.db.
		Not("index_version = ?", indexVersion).
		Order("height").
		First(result).Error

	return result, checkErr(err)
}

// FindMostRecentByDifferentIndexVersion returns the most recent syncable with different index version
func (s SyncablesStore) FindMostRecentByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error) {
	result := &model.Syncable{}

	err := s.db.
		Not("index_version = ?", indexVersion).
		Order("height desc").
		First(result).Error

	return result, checkErr(err)
}

// CreateOrUpdate creates a new syncable or updates an existing one
func (s SyncablesStore) CreateOrUpdate(val *model.Syncable) error {
	existing, err := s.FindByHeight(val.Height)
	if err != nil {
		if err == ErrNotFound {
			return s.Create(val)
		}
		return err
	}
	return s.Update(existing)
}

// CreateOrUpdate creates a new syncable or updates an existing one
func (s SyncablesStore) SetProcessedAtForRange(reportID types.ID, startHeight int64, endHeight int64) error {
	err := s.db.
		Exec("UPDATE syncables SET report_id = ?, processed_at = NULL WHERE height >= ? AND height <= ?", reportID, startHeight, endHeight).
		Error

	return checkErr(err)
}
