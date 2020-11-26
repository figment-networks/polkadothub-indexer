package psql

import (
	"github.com/figment-networks/indexing-engine/store/bulk"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store/psql/queries"

	"github.com/jinzhu/gorm"
)

func NewAccountEraSeqStore(db *gorm.DB) *AccountEraSeqStore {
	return &AccountEraSeqStore{scoped(db, model.AccountEraSeq{})}
}

// AccountEraSeqStore handles operations on accounts
type AccountEraSeqStore struct {
	baseStore
}

// BulkUpsert imports new records and updates existing ones
func (s AccountEraSeqStore) BulkUpsert(records []model.AccountEraSeq) error {
	var err error

	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}

		err = s.Import(queries.AccountEraSeqInsert, j-i, func(k int) bulk.Row {
			r := records[i+k]
			return bulk.Row{
				r.Era,
				r.StartHeight,
				r.EndHeight,
				r.Time,
				r.StashAccount,
				r.ControllerAccount,
				r.ValidatorStashAccount,
				r.ValidatorControllerAccount,
				r.Stake.String(),
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FindByHeight finds account era sequences by era
func (s AccountEraSeqStore) FindByEra(era int64) ([]model.AccountEraSeq, error) {
	q := model.AccountEraSeq{
		EraSequence: &model.EraSequence{
			Era: era,
		},
	}
	var result []model.AccountEraSeq

	err := s.db.
		Where(&q).
		Find(&result).
		Error

	return result, checkErr(err)
}

// FindLastByStashAccount finds last account era sequences for given stash account
func (s AccountEraSeqStore) FindLastByStashAccount(stashAccount string) ([]model.AccountEraSeq, error) {
	rows, err := s.db.
		Raw(queries.AccountEraSeqFindLastByStash, stashAccount, stashAccount).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.AccountEraSeq
	for rows.Next() {
		var row model.AccountEraSeq
		if err := s.db.ScanRows(rows, &row); err != nil {
			return nil, err
		}
		res = append(res, row)
	}
	return res, nil
}

// FindLastByValidatorStashAccount finds last account era sequences for given validator stash account
func (s AccountEraSeqStore) FindLastByValidatorStashAccount(validatorStashAccount string) ([]model.AccountEraSeq, error) {
	rows, err := s.db.
		Raw(queries.AccountEraSeqFindLastByValidatorStash, validatorStashAccount, validatorStashAccount).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.AccountEraSeq
	for rows.Next() {
		var row model.AccountEraSeq
		if err := s.db.ScanRows(rows, &row); err != nil {
			return nil, err
		}
		res = append(res, row)
	}
	return res, nil
}
