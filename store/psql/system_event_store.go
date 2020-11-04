package psql

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/jinzhu/gorm"
)

func NewSystemEventsStore(db *gorm.DB) *SystemEventStore {
	return &SystemEventStore{scoped(db, model.SystemEvent{})}
}

// SystemEventStore handles operations on syncables
type SystemEventStore struct {
	baseStore
}

// CreateOrUpdate creates a new system event or updates an existing one
func (s SystemEventStore) CreateOrUpdate(val *model.SystemEvent) error {
	existing, err := s.findUnique(val.Height, val.Actor, val.Kind)
	if err != nil {
		if err == store.ErrNotFound {
			return s.Create(val)
		}
		return err
	}

	existing.Update(*val)

	return s.Update(existing)
}

func (s SystemEventStore) findUnique(height int64, address string, kind model.SystemEventKind) (*model.SystemEvent, error) {
	q := model.SystemEvent{
		Height: height,
		Actor:  address,
		Kind:   kind,
	}

	var result model.SystemEvent
	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}
