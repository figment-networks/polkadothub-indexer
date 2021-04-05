package psql

import (
	"github.com/jinzhu/gorm"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
)

func NewTransactionSeqStore(db *gorm.DB) *TransactionSeqStore {
	return &TransactionSeqStore{scoped(db, model.TransactionSeq{})}
}

// TransactionSeqStore handles operations on balance events
type TransactionSeqStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s TransactionSeqStore) BulkUpsert(records []model.TransactionSeq) error {
	return s.Import(queries.TransactionSeqInsert, len(records), func(i int) bulk.Row {
		r := records[i]
		return bulk.Row{
			r.Height,
			r.Time,
			r.Index,
			r.Hash,
			r.Method,
			r.Section,
		}
	})
}

// GetTransactionsByTransactionKind gets transactions by kind
func (s TransactionSeqStore) GetTransactionsByTransactionKind(kind model.TransactionKind, start, end int64) ([]model.TransactionSeq, error) {
	var results []model.TransactionSeq

	tx := s.db.
		Where("method = ? AND section = ?", kind.Method, kind.Section)

	if start > 0 {
		tx = tx.Where("height >= ?", start)
	}

	if end > 0 {
		tx = tx.Where("height <= ?", end)
	}

	err := tx.Find(&results).
		Error

	return results, checkErr(err)
}
