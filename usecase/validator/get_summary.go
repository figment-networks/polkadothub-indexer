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

func (uc *getSummaryUseCase) Execute(interval types.SummaryInterval, period string, stashAccount string) (interface{}, error) {
	if stashAccount == "" {
		return uc.validatorSummaryDb.FindSummary(interval, period)
	}
	return uc.validatorSummaryDb.FindSummaryByStashAccount(stashAccount, interval, period)
}
