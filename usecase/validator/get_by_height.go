package validator

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/store"

	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	syncablesDb store.Syncables
	validatorDb store.Validators
}

func NewGetByHeightUseCase(syncablesDb store.Syncables, validatorDb store.Validators) *getByHeightUseCase {
	return &getByHeightUseCase{
		syncablesDb: syncablesDb,
		validatorDb: validatorDb,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return SeqListView{}, err
	}
	lastH := mostRecentSynced.Height

	// Show last synced height, if not provided
	if height == nil {
		height = &lastH
	}

	if *height > lastH {
		return SeqListView{}, errors.New("height is not indexed yet")
	}

	sessionSeqs, err := uc.validatorDb.FindSessionSeqsByHeight(*height)
	if len(sessionSeqs) == 0 || err != nil {
		syncable, err := uc.syncablesDb.FindLastInSessionForHeight(*height)
		if err != nil {
			return SeqListView{}, err
		}

		indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.db, uc.client)
		if err != nil {
			return SeqListView{}, err
		}

		ctx := context.Background()
		payload, err := indexingPipeline.Run(ctx, indexer.RunConfig{
			Height:           syncable.Height,
			DesiredTargetIDs: []int64{indexer.TargetIndexValidatorSessionSequences},
			Dry:              true,
		})
		if err != nil {
			return SeqListView{}, err
		}

		sessionSeqs = append(payload.NewValidatorSessionSequences, payload.UpdatedValidatorSessionSequences...)
	}

	eraSeqs, err := uc.validatorDb.FindEraSeqsByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return SeqListView{}, err
	}

	return ToSeqListView(sessionSeqs, eraSeqs), nil
}
