package block

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getByHeightUseCase struct {
	client *client.Client

	syncablesDb store.Syncables
}

func NewGetByHeightUseCase(c *client.Client, syncablesDb store.Syncables) *getByHeightUseCase {
	return &getByHeightUseCase{
		syncablesDb: syncablesDb,
		client:      c,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (*DetailsView, error) {
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

	res, err := uc.client.Block.GetByHeight(*height)
	if err != nil {
		return nil, err
	}

	return ToDetailsView(res), nil
}
