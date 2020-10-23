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

func (uc *getSummaryUseCase) Execute(interval types.SummaryInterval, period string, stashAccount string) (summaryListView, error) {
	var err error
	var summaries []store.ValidatorSummaryRow

	lastIndexedSession, err := uc.db.Syncables.FindLastEndOfSession()
	if err != nil && err != store.ErrNotFound {
		return summaryListView{}, err
	}

	lastIndexedEra, err := uc.db.Syncables.FindLastEndOfEra()
	if err != nil && err != store.ErrNotFound {
		return summaryListView{}, err
	}

	if stashAccount == "" {
		summaries, err = uc.db.ValidatorSummary.FindSummary(interval, period)
	} else {
		summaries, err = uc.db.ValidatorSummary.FindSummaryByStashAccount(stashAccount, interval, period)
	}

	if err != nil {
		return summaryListView{}, err
	}

	return toSummaryListView(summaries, lastIndexedSession, lastIndexedEra), nil
}
