package store

import "github.com/figment-networks/polkadothub-indexer/model"

type EventSeq interface {
	baseStore
	FindByHeightAndIndex(height int64, index int64) (*model.EventSeq, error)
	FindBalanceDeposits(address string) ([]model.EventSeq, error)
	FindBalanceTransfers(address string) ([]model.EventSeq, error)
	FindBonded(address string) ([]model.EventSeq, error)
	FindUnbonded(address string) ([]model.EventSeq, error)
	FindWithdrawn(address string) ([]model.EventSeq, error)
}
