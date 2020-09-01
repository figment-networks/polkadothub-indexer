package account

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getDetailsUseCase struct {
	db     *store.Store
	client *client.Client
}

func NewGetDetailsUseCase(db *store.Store, c *client.Client) *getDetailsUseCase {
	return &getDetailsUseCase{
		db:     db,
		client: c,
	}
}

func (uc *getDetailsUseCase) Execute(address string) (*DetailsView, error) {
	identity, err := uc.client.Account.GetIdentity(address)
	if err != nil {
		return nil, err
	}

	eraLimit := int64(1)
	accountEraSeqs, err := uc.db.AccountEraSeq.FindLastByStashAccount(address, eraLimit)
	if err != nil {
		return nil, err
	}

	balanceTransfers, err := uc.db.EventSeq.FindBalanceTransfers(address)
	if err != nil {
		return nil, err
	}

	balanceDeposits, err := uc.db.EventSeq.FindBalanceDeposits(address)
	if err != nil {
		return nil, err
	}

	bonded, err := uc.db.EventSeq.FindBonded(address)
	if err != nil {
		return nil, err
	}

	unbonded, err := uc.db.EventSeq.FindUnbonded(address)
	if err != nil {
		return nil, err
	}

	withdrawn, err := uc.db.EventSeq.FindWithdrawn(address)
	if err != nil {
		return nil, err
	}

	return ToDetailsView(address, identity.GetIdentity(), accountEraSeqs, balanceTransfers, balanceDeposits, bonded, unbonded, withdrawn)
}

