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

// FindByActor returns system events by actor
func (s SystemEventStore) FindByActor(actorAddress string, kind *model.SystemEventKind, minHeight *int64) ([]model.SystemEvent, error) {
	var result []model.SystemEvent
	q := model.SystemEvent{}
	if kind != nil {
		q.Kind = *kind
	}

	statement := s.db.
		Where("actor = ?", actorAddress).
		Where(&q)

	if minHeight != nil {
		statement = statement.Where("height > ?", *minHeight)
	}

	err := statement.
		Find(&result).
		Error

	return result, checkErr(err)
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
