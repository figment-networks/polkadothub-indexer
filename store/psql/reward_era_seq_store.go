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
				r.TxHash,
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// MarkAllClaimed updates all rewards for validatorStash and era as claimed. Returns error if nothing updates
func (s RewardEraSeqStore) MarkAllClaimed(validatorStash string, era int64, txHash string) error {
	// Update with conditions
	result := s.db.Model(&model.RewardEraSeq{}).
		Where("validator_stash_account = ? AND era = ?", validatorStash, era).
		Update("claimed", true).
		Update("tx_hash", txHash)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errExpectedUpdate
	}

	return nil
}

// GetAll Gets all rewards for given stash
func (s RewardEraSeqStore) GetAll(stash string, start, end int64) ([]model.RewardEraSeq, error) {
	tx := s.db.
		Table(model.RewardEraSeq{}.TableName()).
		Select("*").
		Where("stash_account = ?", stash).
		Order("era")

	if end != 0 {
		tx = tx.Where("era <= ?", end)
	}
	if start != 0 {
		tx = tx.Where("era >= ?", start)
	}

	var res []model.RewardEraSeq
	return res, tx.Find(&res).Error
}

// GetCount returns record count for given validatorStash at era
func (s RewardEraSeqStore) GetCount(validatorStash string, era int64) (int64, error) {
	var count int64
	err := s.db.
		Table(model.RewardEraSeq{}.TableName()).
		Where("validator_stash_account = ? AND era = ?", validatorStash, era).
		Count(&count).
		Error

	if gorm.IsRecordNotFoundError(err) {
		return 0, nil
	}

	return count, err
}

// GetByStashAndEra returns for given stash for validatorStash and era
func (s *RewardEraSeqStore) GetByStashAndEra(validatorStash, stash string, era int64) (model.RewardEraSeq, error) {
	var result model.RewardEraSeq
	err := s.db.
		Table(model.RewardEraSeq{}.TableName()).
		Where("validator_stash_account = ? AND stash_account = ? AND era = ?", validatorStash, stash, era).
		Count(&result).
		Error

	return result, checkErr(err)
}
