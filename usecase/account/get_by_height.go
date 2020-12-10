package account

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getByHeightUseCase struct {
	client *client.Client

	syncablesDb store.FindMostRecenter
}

func NewGetByHeightUseCase(c *client.Client, syncablesDb store.FindMostRecenter) *getByHeightUseCase {
	return &getByHeightUseCase{
		client: c,

		syncablesDb: syncablesDb,
	}
}

func (uc *getByHeightUseCase) Execute(address string, height *int64) (*HeightDetailsView, error) {
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

	res, err := uc.client.Account.GetByHeight(address, *height)
	if err != nil {
		return nil, err
	}

	return ToHeightDetailsView(res.GetAccount()), nil
}
