package psql

import (
	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
	"github.com/jinzhu/gorm"
)

func NewRewardEraSeqStore(db *gorm.DB) *RewardEraSeqStore {
	return &RewardEraSeqStore{scoped(db, model.RewardEraSeq{})}
}

// RewardEraSeqStore handles operations on rewardEraSeq
type RewardEraSeqStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s RewardEraSeqStore) BulkUpsert(records []model.RewardEraSeq) error {
	var err error

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}

		err = s.Import(queries.RewardEraSeqInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				r.Era,
				r.StartHeight,
				r.EndHeight,
				r.Time,
				r.StashAccount,
				r.ValidatorStashAccount,
				r.Amount,
				r.Kind,
				r.Claimed,
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
