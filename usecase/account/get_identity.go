package account

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getIdentityUseCase struct {
	db     *store.Store
	client *client.Client
}

func NewGetIdentityUseCase(db *store.Store, c *client.Client) *getIdentityUseCase {
	return &getIdentityUseCase{
		db:     db,
		client: c,
	}
}

func (uc *getIdentityUseCase) Execute(address string) (*IdentityDetailsView, error) {
	res, err := uc.client.Account.GetIdentity(address)
	if err != nil {
		return nil, err
	}

	return ToIdentityDetailsView(res.GetIdentity()), nil
}

