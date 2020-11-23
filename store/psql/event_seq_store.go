package psql

import (
	"time"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"

	"github.com/jinzhu/gorm"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewEventSeqStore(db *gorm.DB) *EventSeqStore {
	return &EventSeqStore{scoped(db, model.EventSeq{})}
}

// EventSeqStore handles operations on events
type EventSeqStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s EventSeqStore) BulkUpsert(records []model.EventSeq) error {
	var err error

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}

		err = s.Import(queries.EventSeqInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				r.Height,
				r.Time,
				r.Index,
				r.ExtrinsicIndex,
				r.Data,
				r.Phase,
				r.Method,
				r.Section,
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FindByHeightAndStashAccount finds event by height and index
func (s EventSeqStore) FindByHeightAndIndex(height int64, index int64) (*model.EventSeq, error) {
	q := model.EventSeq{
		Sequence: &model.Sequence{
			Height: height,
		},
		Index: index,
	}
	var result model.EventSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindByHeight finds event sequences by height
func (s EventSeqStore) FindByHeight(height int64) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Sequence: &model.Sequence{
			Height: height,
		},
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBalanceTransfers finds balance transfers event sequences for given address
func (s EventSeqStore) FindBalanceTransfers(address string) ([]model.EventSeqWithTxHash, error) {
	return s.findForEventSeqWithTxHashQuery("balances", "Transfer", address)
}

// FindBalanceDeposits finds balance deposits event sequences for given address
func (s EventSeqStore) FindBalanceDeposits(address string) ([]model.EventSeqWithTxHash, error) {
	return s.findForEventSeqWithTxHashQuery("balances", "Deposit", address)
}

// FindBonded finds bonded event sequences for given address
func (s EventSeqStore) FindBonded(address string) ([]model.EventSeqWithTxHash, error) {
	return s.findForEventSeqWithTxHashQuery("staking", "Bonded", address)
}

// FindUnbonded finds unbonded event sequences for given address
func (s EventSeqStore) FindUnbonded(address string) ([]model.EventSeqWithTxHash, error) {
	return s.findForEventSeqWithTxHashQuery("staking", "Unbonded", address)
}

// FindWithdrawn finds withdrawn event sequences for given address
func (s EventSeqStore) FindWithdrawn(address string) ([]model.EventSeqWithTxHash, error) {
	return s.findForEventSeqWithTxHashQuery("staking", "Withdrawn", address)
}

// FindMostRecent finds most recent event session sequence
func (s *EventSeqStore) FindMostRecent() (*model.EventSeq, error) {
	eventSeq := &model.EventSeq{}
	if err := findMostRecent(s.db, "time", eventSeq); err != nil {
		return nil, err
	}
	return eventSeq, nil
}

// DeleteOlderThan deletes event sequence older than given threshold
func (s *EventSeqStore) DeleteOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.EventSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}

func (s EventSeqStore) findForEventSeqWithTxHashQuery(section, method, address string) ([]model.EventSeqWithTxHash, error) {
	var result []model.EventSeqWithTxHash

	tx := s.db
	if method == "Transfer" {
		tx = s.db.
			Raw(queries.EventSeqWithTxHashForSrcAndTarget, section, method, address, address)
	} else {
		tx = s.db.
			Raw(queries.EventSeqWithTxHashForSrc, section, method, address)
	}

	rows, err := tx.Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		event := model.EventSeqWithTxHash{}
		if err := rows.Scan(&event.Height, &event.Method, &event.Section, &event.Data, &event.TxHash); err != nil {
			return nil, err
		}
		result = append(result, event)
	}

	return result, nil
}
