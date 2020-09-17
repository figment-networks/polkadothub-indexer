package transaction

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	db     store.Store
	client *client.Client
}

func NewGetByHeightUseCase(db store.Store, c *client.Client) *getByHeightUseCase {
	return &getByHeightUseCase{
		db:     db,
		client: c,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (*ListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.db.GetSyncables().FindMostRecent()
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
