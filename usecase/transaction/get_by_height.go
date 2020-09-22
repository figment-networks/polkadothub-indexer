package transaction

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	client *client.Client

	syncablesDb store.Syncables
}

func NewGetByHeightUseCase(c *client.Client, syncablesDb store.Syncables) *getByHeightUseCase {
	return &getByHeightUseCase{
		client: c,

		syncablesDb: syncablesDb,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (*ListView, error) {
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

	res, err := uc.client.Transaction.GetByHeight(*height)
	if err != nil {
		return nil, err
	}

	return ToListView(res.GetTransactions()), nil
}
