package validator

import (
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type getSummaryUseCase struct {
	db *store.Store
}

func NewGetSummaryUseCase(db *store.Store) *getSummaryUseCase {
	return &getSummaryUseCase{
		db: db,
	}
}

func (uc *getSummaryUseCase) Execute(interval types.SummaryInterval, period string, stashAccount string) (interface{}, error) {
	if stashAccount == "" {
		return uc.db.ValidatorSummary.FindSummary(interval, period)
	}
	return uc.db.ValidatorSummary.FindSummaryByStashAccount(stashAccount, interval, period)
}

