package psql

import (
	"fmt"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/jinzhu/gorm"
)

// baseStore implements generic store operations
type baseStore struct {
	db    *gorm.DB
	model interface{}
}

// Create creates a new record. Must pass a pointer.
func (s baseStore) Create(record interface{}) error {
	err := s.db.Create(record).Error
	return checkErr(err)
}

// Update updates the existing record. Must pass a pointer.
func (s baseStore) Update(record interface{}) error {
	err := s.db.Save(record).Error
	return checkErr(err)
}

// Import imports records in bulk
func (s baseStore) Import(query string, rows int, fn bulk.RowFunc) error {
	return bulk.Import(s.db, query, rows, fn)
}

// Save saves record to database
func (s baseStore) Save(record interface{}) error {
	return s.db.Save(record).Error
}

// Truncate removes all records from the table
func (s baseStore) Truncate() error {
	return s.db.Delete(s.model).Error
}

// DeleteByHeight removes all records associated with a height
func (s baseStore) DeleteByHeight(height int64) error {
	return s.db.Delete(s.model, "height = ?", height).Error
}

func scoped(conn *gorm.DB, m interface{}) baseStore {
	return baseStore{conn, m}
}

func isNotFound(err error) bool {
	return gorm.IsRecordNotFoundError(err) || err == store.ErrNotFound
}

func findBy(db *gorm.DB, dst interface{}, key string, value interface{}) error {
	return db.
		Model(dst).
		Where(fmt.Sprintf("%s = ?", key), value).
		First(dst).
		Error
}

func findMostRecent(db *gorm.DB, orderField string, record interface{}) error {
	return db.
		Order(fmt.Sprintf("%s DESC", orderField)).
		Take(record).
		Error
}

func checkErr(err error) error {
	if gorm.IsRecordNotFoundError(err) {
		return store.ErrNotFound
	}
	return err
}
