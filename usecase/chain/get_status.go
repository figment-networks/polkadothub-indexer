package chain

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getStatusUseCase struct {
	db     *store.Store
	client *client.Client
}

func NewGetStatusUseCase(db *store.Store, c *client.Client) *getStatusUseCase {
	return &getStatusUseCase{
		db:     db,
		client: c,
	}
}

func (uc *getStatusUseCase) Execute(ctx context.Context) (*DetailsView, error) {
	var lastEraHeight, lastSessionHeight int64

	mostRecentSyncable, err := uc.db.Syncables.FindMostRecent()
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	if mostRecentSyncable != nil {
		lastSessionSyncable, err := uc.db.Syncables.FindLastInSession(mostRecentSyncable.Session)
		if err != nil && err != store.ErrNotFound {
			return nil, err
		}
		if lastSessionSyncable != nil {
			lastSessionHeight = lastSessionSyncable.Height
		}

		lastEraSyncable, err := uc.db.Syncables.FindLastInSession(mostRecentSyncable.Era)
		if err != nil && err != store.ErrNotFound {
			return nil, err
		}
		if lastEraSyncable != nil {
			lastSessionHeight = lastEraSyncable.Height
		}
	}

	getHeadRes, err := uc.client.Chain.GetHead()
	if err != nil {
		return nil, err
	}

	getStatusRes, err := uc.client.Chain.GeStatus()
	if err != nil {
		return nil, err
	}

	return ToDetailsView(mostRecentSyncable, getHeadRes, getStatusRes, lastSessionHeight, lastEraHeight), nil
}
