package psql

import (
	"errors"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
	"github.com/jinzhu/gorm"
)

var (
	errExpectedUpdate = errors.New("no rewards were updated")
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

// MarkAllClaimed updates all rewards for validatorStash and era as claimed. Returns error if nothing updates
func (s RewardsStore) MarkAllClaimed(validatorStash string, era int64) error {
	// Update with conditions
	result := s.db.Model(&model.Reward{}).
		Where("validator_stash_account = ? AND era = ?", validatorStash, era).
		Update("claimed", true)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errExpectedUpdate
	}

	return nil
}

// GetAll Gets all rewards for given stash
func (s RewardsStore) GetAll(stash string, start, end int64) ([]model.Reward, error) {
	tx := s.db.
		Model(&model.Reward{}).
		Select("*").
		Where("stash_account = ?", stash).
		Order("era")

	if end != 0 {
		tx = tx.Where("era <= ?", end)
	}
	if start != 0 {
		tx = tx.Where("era >= ?", start)
	}

	var res []model.Reward
	return res, tx.Find(&res).Error
}
