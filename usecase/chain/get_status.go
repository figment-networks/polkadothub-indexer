package chain

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getStatusUseCase struct {
	client *client.Client

	syncablesDb store.Syncables
}

func NewGetStatusUseCase(c *client.Client, syncablesDb store.Syncables) *getStatusUseCase {
	return &getStatusUseCase{
		syncablesDb: syncablesDb,
		client:      c,
	}
}

func (uc *getStatusUseCase) Execute(ctx context.Context) (*DetailsView, error) {
	mostRecentSyncable, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		if err != store.ErrNotFound {
			return nil, err
		} else {
			mostRecentSyncable = nil
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

	return ToDetailsView(mostRecentSyncable, getHeadRes, getStatusRes), nil
}
