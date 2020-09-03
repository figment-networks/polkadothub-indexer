package store

import (
	"github.com/jinzhu/gorm"
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
)

func NewAccountEraSeqStore(db *gorm.DB) *AccountEraSeqStore {
	return &AccountEraSeqStore{scoped(db, model.AccountEraSeq{})}
}

// AccountEraSeqStore handles operations on accounts
type AccountEraSeqStore struct {
	baseStore
}

// CreateIfNotExists creates the account if it does not exist
func (s AccountEraSeqStore) CreateIfNotExists(account *model.AccountEraSeq) error {
	_, err := s.FindByHeight(account.Era)
	if isNotFound(err) {
		return s.Create(account)
	}
	return nil
}

// FindByHeightAndStashAccount finds account by height and stash account
func (s AccountEraSeqStore) FindByHeightAndStashAccounts(height int64, stash string, validatorStash string) (*model.AccountEraSeq, error) {
	q := model.AccountEraSeq{
		StashAccount: stash,
		ValidatorStashAccount: validatorStash,
	}
	var result model.AccountEraSeq

	err := s.db.
		Where(&q).
		Where("start_height >= ? AND end_height <= ?", height, height).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindByEraAndStashAccount finds account by era and stash account
func (s AccountEraSeqStore) FindByEraAndStashAccount(era int64, stash string) (*model.AccountEraSeq, error) {
	q := model.AccountEraSeq{
		EraSequence: &model.EraSequence{
			Era: era,
		},
		StashAccount: stash,
	}
	var result model.AccountEraSeq

	err := s.db.
		Where(&q).
		First(&result).
		Error

	return &result, checkErr(err)
}

// FindByHeight finds account era sequences by height
func (s AccountEraSeqStore) FindByHeight(h int64) ([]model.AccountEraSeq, error) {
	var result []model.AccountEraSeq

	err := s.db.
		Where("start_height >= ? AND end_height <= ?", h, h).
		Find(&result).
		Error

	return result, checkErr(err)
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

// FindMostRecent finds most recent account era sequence
func (s *AccountEraSeqStore) FindMostRecent() (*model.AccountEraSeq, error) {
	accountSeq := &model.AccountEraSeq{}
	if err := findMostRecent(s.db, "time", accountSeq); err != nil {
		return nil, err
	}
	return accountSeq, nil
}

// FindLastByStashAccount finds last account era sequences for given stash account and era limit
func (s AccountEraSeqStore) FindLastByStashAccount(stashAccount string, eraLimit int64) ([]model.AccountEraSeq, error) {
	rows, err := s.db.
		Raw(findLastAccountEraSeqByStashQuery, stashAccount, stashAccount, eraLimit).
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

// FindLastByValidatorStashAccount finds last account era sequences for given validator stash account and era limit
func (s AccountEraSeqStore) FindLastByValidatorStashAccount(validatorStashAccount string, eraLimit int64) ([]model.AccountEraSeq, error) {
	rows, err := s.db.
		Raw(findLastAccountEraSeqByValidatorStashQuery, validatorStashAccount, validatorStashAccount, eraLimit).
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

// DeleteOlderThan deletes account sequence older than given threshold
func (s *AccountEraSeqStore) DeleteOlderThan(purgeThreshold time.Time) (*int64, error) {
	tx := s.db.
		Unscoped().
		Where("time < ?", purgeThreshold).
		Delete(&model.AccountEraSeq{})

	if tx.Error != nil {
		return nil, checkErr(tx.Error)
	}

	return &tx.RowsAffected, nil
}
