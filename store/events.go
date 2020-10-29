package store

import "github.com/figment-networks/polkadothub-indexer/model"

type EventSeq interface {
	baseStore
	FindAllByHeightAndIndex(height int64, indexes []int64) (map[int64]*model.EventSeq, error)
	FindByHeightAndIndex(height int64, index int64) (*model.EventSeq, error)
	FindBalanceDeposits(address string) ([]model.EventSeqWithTxHash, error)
	FindBalanceTransfers(address string) ([]model.EventSeqWithTxHash, error)
	FindBonded(address string) ([]model.EventSeqWithTxHash, error)
	FindUnbonded(address string) ([]model.EventSeqWithTxHash, error)
	FindWithdrawn(address string) ([]model.EventSeqWithTxHash, error)
}
