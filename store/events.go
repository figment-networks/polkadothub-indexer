package store

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
)

type EventSeq interface {
	BulkUpsert(records []model.EventSeq) error
	FindByHeightAndIndex(height int64, index int64) (*model.EventSeq, error)
	FindBalanceDeposits(address string) ([]model.EventSeqWithTxHash, error)
	FindBalanceTransfers(address string) ([]model.EventSeqWithTxHash, error)
	FindBonded(address string) ([]model.EventSeqWithTxHash, error)
	FindRewardsForTimePeriod(address string, start, end time.Time) ([]model.EventSeq, error)
	FindUnbonded(address string) ([]model.EventSeqWithTxHash, error)
	FindWithdrawn(address string) ([]model.EventSeqWithTxHash, error)
}
