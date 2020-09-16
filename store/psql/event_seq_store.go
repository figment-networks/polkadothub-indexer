package store

import (
	"github.com/jinzhu/gorm"
	"time"

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
func (s EventSeqStore) FindBalanceTransfers(address string) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Section: "balances",
		Method:  "Transfer",
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Or("event_sequences.data->1->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBalanceDeposits finds balance deposits event sequences for given address
func (s EventSeqStore) FindBalanceDeposits(address string) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Section: "balances",
		Method:  "Deposit",
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBonded finds bonded event sequences for given address
func (s EventSeqStore) FindBonded(address string) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Section: "staking",
		Method:  "Bonded",
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBonded finds unbonded event sequences for given address
func (s EventSeqStore) FindUnbonded(address string) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Section: "staking",
		Method:  "Unbonded",
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBonded finds withdrawn event sequences for given address
func (s EventSeqStore) FindWithdrawn(address string) ([]model.EventSeq, error) {
	q := model.EventSeq{
		Section: "staking",
		Method:  "Withdrawn",
	}
	var result []model.EventSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
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
