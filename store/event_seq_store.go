package store

import (
	"time"

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
	var result []model.EventSeqWithTxHash

	rows, err := s.db.
		Table(model.EventSeq{}.TableName()).
		Select(eventSeqWithTxHashQuerySelect).
		Joins("INNER JOIN transaction_sequences ON transaction_sequences.height = event_sequences.height AND transaction_sequences.index = event_sequences.extrinsic_index ").
		Where("event_sequences.section='balances' AND event_sequences.method='Transfer'").
		Where("event_sequences.data->0->>'value' = ?", address).
		Or("event_sequences.data->1->>'value' = ?", address).
		Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		tmp := model.EventSeqWithTxHash{}
		rows.Scan(&tmp.Height, &tmp.Method, &tmp.Section, &tmp.Data, &tmp.TxHash)
		result = append(result, tmp)
	}

	return result, nil
}

const (
	eventSeqWithTxHashQuerySelect = `
	event_sequences.height,
	event_sequences.method,
	event_sequences.section,
	event_sequences.data,
	transaction_sequences.hash
`
)

// FindBalanceDeposits finds balance deposits event sequences for given address
func (s EventSeqStore) FindBalanceDeposits(address string) ([]model.EventSeqWithTxHash, error) {
	var result []model.EventSeqWithTxHash

	rows, err := s.db.
		Table(model.EventSeq{}.TableName()).
		Select(eventSeqWithTxHashQuerySelect).
		Joins("INNER JOIN transaction_sequences ON transaction_sequences.height = event_sequences.height AND transaction_sequences.index = event_sequences.extrinsic_index ").
		Where("event_sequences.section='balances' AND event_sequences.method='Deposit'").
		Where("event_sequences.data->0->>'value' = ?", address).
		Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		tmp := model.EventSeqWithTxHash{}
		rows.Scan(&tmp.Height, &tmp.Method, &tmp.Section, &tmp.Data, &tmp.TxHash)
		result = append(result, tmp)
	}

	return result, nil
}

// FindBonded finds bonded event sequences for given address
func (s EventSeqStore) FindBonded(address string) ([]model.EventSeqWithTxHash, error) {
	var result []model.EventSeqWithTxHash

	rows, err := s.db.
		Table(model.EventSeq{}.TableName()).
		Select(eventSeqWithTxHashQuerySelect).
		Joins("INNER JOIN transaction_sequences ON transaction_sequences.height = event_sequences.height AND transaction_sequences.index = event_sequences.extrinsic_index ").
		Where("event_sequences.section='staking' AND event_sequences.method='Bonded'").
		Where("event_sequences.data->0->>'value' = ?", address).
		Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		tmp := model.EventSeqWithTxHash{}
		rows.Scan(&tmp.Height, &tmp.Method, &tmp.Section, &tmp.Data, &tmp.TxHash)
		result = append(result, tmp)
	}

	return result, nil
}

// FindUnbonded finds unbonded event sequences for given address
func (s EventSeqStore) FindUnbonded(address string) ([]model.EventSeqWithTxHash, error) {
	var result []model.EventSeqWithTxHash

	rows, err := s.db.
		Table(model.EventSeq{}.TableName()).
		Select(eventSeqWithTxHashQuerySelect).
		Joins("INNER JOIN transaction_sequences ON transaction_sequences.height = event_sequences.height AND transaction_sequences.index = event_sequences.extrinsic_index ").
		Where("event_sequences.section='staking' AND event_sequences.method='Unbonded'").
		Where("event_sequences.data->0->>'value' = ?", address).
		Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		tmp := model.EventSeqWithTxHash{}
		rows.Scan(&tmp.Height, &tmp.Method, &tmp.Section, &tmp.Data, &tmp.TxHash)
		result = append(result, tmp)
	}

	return result, nil
}

// FindWithdrawn finds withdrawn event sequences for given address
func (s EventSeqStore) FindWithdrawn(address string) ([]model.EventSeqWithTxHash, error) {
	var result []model.EventSeqWithTxHash

	rows, err := s.db.
		Table(model.EventSeq{}.TableName()).
		Select(eventSeqWithTxHashQuerySelect).
		Joins("INNER JOIN transaction_sequences ON transaction_sequences.height = event_sequences.height AND transaction_sequences.index = event_sequences.extrinsic_index ").
		Where("event_sequences.section='staking' AND event_sequences.method='Withdrawn'").
		Where("event_sequences.data->0->>'value' = ?", address).
		Rows()

	if err != nil {
		return result, checkErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		tmp := model.EventSeqWithTxHash{}
		rows.Scan(&tmp.Height, &tmp.Method, &tmp.Section, &tmp.Data, &tmp.TxHash)
		result = append(result, tmp)
	}

	return result, nil
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
