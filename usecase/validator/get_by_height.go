package validator

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
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

func (uc *getByHeightUseCase) Execute(height *int64) (*SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.db.Syncables.FindMostRecent()
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

	validatorSessionSequences, err := uc.db.ValidatorSessionSeq.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	validatorEraSequences, err := uc.db.ValidatorEraSeq.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	return ToSeqListView(validatorSessionSequences, validatorEraSequences), nil
}
