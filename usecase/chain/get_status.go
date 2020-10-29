package chain

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-proxy/grpc/chain/chainpb"
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

func (uc *getStatusUseCase) Execute(includeChainStatus bool) (*DetailsView, error) {
	var lastEraHeight, lastSessionHeight int64
	var getHeadRes *chainpb.GetHeadResponse
	var getStatusRes *chainpb.GetStatusResponse

	mostRecentSyncable, err := uc.syncablesDb.FindMostRecent()
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	if mostRecentSyncable != nil {
		lastSessionSyncable, err := uc.syncablesDb.FindLastInSession(mostRecentSyncable.Session - 1)
		if err != nil && err != store.ErrNotFound {
			return nil, err
		}
		if lastSessionSyncable != nil {
			lastSessionHeight = lastSessionSyncable.Height
		}

		lastEraSyncable, err := uc.syncablesDb.FindLastInEra(mostRecentSyncable.Era - 1)
		if err != nil && err != store.ErrNotFound {
			return nil, err
		}
		if lastEraSyncable != nil {
			lastEraHeight = lastEraSyncable.Height
		}
	}

	if includeChainStatus {
		getHeadRes, err = uc.client.Chain.GetHead()
		if err != nil {
			return nil, err
		}

		getStatusRes, err = uc.client.Chain.GeStatus()
		if err != nil {
			return nil, err
		}
	}

	return ToDetailsView(mostRecentSyncable, getHeadRes, getStatusRes, lastSessionHeight, lastEraHeight), nil
}
