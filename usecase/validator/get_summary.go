package validator

import (
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type getSummaryUseCase struct {
	validatorSummaryDb store.ValidatorSummary
}

func NewGetSummaryUseCase(validatorSummaryDb store.ValidatorSummary) *getSummaryUseCase {
	return &getSummaryUseCase{
		validatorSummaryDb: validatorSummaryDb,
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
		summaries, err = uc.validatorSummaryDb.FindSummary(interval, period)
	} else {
		summaries, err = uc.validatorSummaryDb.FindSummaryByStashAccount(stashAccount, interval, period)
	}

	if err != nil {
		return summaryListView{}, err
	}

	return toSummaryListView(summaries, lastIndexedSession, lastIndexedEra), nil
}
