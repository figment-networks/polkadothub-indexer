package psql

import (
	"time"

	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"
	"github.com/jinzhu/gorm"
)

func NewValidatorSeqStore(db *gorm.DB) *ValidatorSeqStore {
	return &ValidatorSeqStore{scoped(db, model.ValidatorSeq{})}
}

// ValidatorSeqStore handles operations on validators
type ValidatorSeqStore struct {
	baseStore
}

// BulkUpsertSeqs imports new records and updates existing ones
func (s ValidatorSeqStore) BulkUpsertSeqs(records []model.ValidatorSeq) error {
	var err error

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}

		err = s.Import(queries.ValidatorSeqInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				r.Height,
				r.Time,
				r.StashAccount,
				r.ActiveBalance.String(),
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FindAllByHeight returns all validators for provided height
func (s ValidatorSeqStore) FindAllByHeight(height int64) ([]model.ValidatorSeq, error) {
	var results []model.ValidatorSeq

	err := s.db.
		Where("height = ?", height).
		Find(&results).
		Error

	return results, checkErr(err)
}

// FindMostRecentSeq finds most recent validator sequences
func (s *ValidatorSeqStore) FindMostRecentSeq() (*model.ValidatorSeq, error) {
	validatorSeq := &model.ValidatorSeq{}
	if err := findMostRecent(s.db, "time", validatorSeq); err != nil {
		return nil, err
	}
	return validatorSeq, nil
}

// DeleteSeqsOlderThan deletes validator sequences older than given threshold
func (s *ValidatorSeqStore) DeleteSeqsOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.ValidatorSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}
