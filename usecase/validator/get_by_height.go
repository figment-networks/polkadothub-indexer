package validator

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client
}

func NewGetByHeightUseCase(cfg *config.Config, db *store.Store, client *client.Client) *getByHeightUseCase {
	return &getByHeightUseCase{
		cfg:    cfg,
		db:     db,
		client: client,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.db.Syncables.FindMostRecent()
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

	sessionSeqs, err := uc.db.ValidatorSessionSeq.FindByHeight(*height)
	if len(sessionSeqs) == 0 || err != nil {
		syncable, err := uc.db.Syncables.FindLastInSessionForHeight(*height)
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

		sessionSeqs = payload.NewValidatorSessionSequences
	}

	eraSeqs, err := uc.db.ValidatorEraSeq.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return SeqListView{}, err
	}

	return ToSeqListView(sessionSeqs, eraSeqs), nil
}
