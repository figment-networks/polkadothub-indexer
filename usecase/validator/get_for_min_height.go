package validator

import (
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

type getForMinHeightUseCase struct {
	syncablesDb    store.Syncables
	validatorAggDb store.ValidatorAgg
}

func NewGetForMinHeightUseCase(syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg) *getForMinHeightUseCase {
	return &getForMinHeightUseCase{
		syncablesDb:    syncablesDb,
		validatorAggDb: validatorAggDb,
	}
}

func (uc *getForMinHeightUseCase) Execute(height *int64) (*AggListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return nil, err
	}
	lastH := mostRecentSynced.Height

	// Show last synced height, if not provided
	if height == nil {
		height = &lastH
	}

	if *height > lastH {
		return nil, errors.New("height is not indexed yet")
	}

	validatorAggs, err := uc.validatorAggDb.GetAllForHeightGreaterThan(*height)
	if err != nil {
		return nil, err
	}

	return ToAggListView(validatorAggs), nil
}
