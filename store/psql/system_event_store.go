package psql

import (
	"time"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
	"github.com/jinzhu/gorm"
)

func NewSystemEventsStore(db *gorm.DB) *SystemEventStore {
	return &SystemEventStore{scoped(db, model.SystemEvent{})}
}

// SystemEventStore handles operations on syncables
type SystemEventStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s SystemEventStore) BulkUpsert(records []model.SystemEvent) error {
	var err error
	t := time.Now()

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}
		err = s.Import(queries.SystemEventInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				t,
				t,
				r.Height,
				r.Time,
				r.Actor,
				r.Kind,
				r.Data,
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
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
