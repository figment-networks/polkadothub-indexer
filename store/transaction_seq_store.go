package store

import (
	"github.com/jinzhu/gorm"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewTransactionSeqStore(db *gorm.DB) *TransactionSeqStore {
	return &TransactionSeqStore{scoped(db, model.TransactionSeq{})}
}

// TransactionSeqStore handles operations on balance events
type TransactionSeqStore struct {
	baseStore
}

// FindByHeightAndIndex finds event by height and index
func (s TransactionSeqStore) FindByHeightAndIndex(height int64, index int64) (*model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Sequence: &model.Sequence{
			Height: height,
		},
		Index: index,
	}
	var result model.TransactionSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindBalanceTransfers finds all balance transfer events for given address
func (s TransactionSeqStore) FindBalanceTransfers(address string) ([]model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Section: "balances",
		Method:  "Transfer",
	}
	var result []model.TransactionSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Or("event_sequences.data->1->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBalanceDeposits finds all balance deposit events for given address
func (s TransactionSeqStore) FindBalanceDeposits(address string) ([]model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Section: "balances",
		Method:  "Deposit",
	}
	var result []model.TransactionSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBonded finds all bonded events for given address
func (s TransactionSeqStore) FindBonded(address string) ([]model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Section: "staking",
		Method:  "Bonded",
	}
	var result []model.TransactionSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindBonded finds all unbonded events for given address
func (s TransactionSeqStore) FindUnbonded(address string) ([]model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Section: "staking",
		Method:  "Unbonded",
	}
	var result []model.TransactionSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindWithdrawn finds all withdrawn events for given address
func (s TransactionSeqStore) FindWithdrawn(address string) ([]model.TransactionSeq, error) {
	q := model.TransactionSeq{
		Section: "staking",
		Method:  "Withdrawn",
	}
	var result []model.TransactionSeq

	err := s.db.
		Where(&q).
		Where("event_sequences.data->0->>'value' = ?", address).
		Find(&result).
		Error

	return result, checkErr(err)
}
