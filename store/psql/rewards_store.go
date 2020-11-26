package psql

import (
	"time"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
	"github.com/jinzhu/gorm"
)

func NewRewardsStore(db *gorm.DB) *RewardsStore {
	return &RewardsStore{scoped(db, model.Reward{})}
}

// RewardsStore handles operations on rewards
type RewardsStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s RewardsStore) BulkUpsert(records []model.Reward) error {
	var err error
	t := time.Now()

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}

		err = s.Import(queries.RewardInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				t,
				t,
				r.Era,
				r.StashAccount,
				r.ValidatorStashAccount,
				r.Amount,
				r.Kind,
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
