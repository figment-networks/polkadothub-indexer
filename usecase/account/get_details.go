package account

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getDetailsUseCase struct {
	client *client.Client

	accountEraSeqDb store.AccountEraSeq
	eventSeqDb      store.EventSeq
}

func NewGetDetailsUseCase(c *client.Client, accountEraSeqDb store.AccountEraSeq, eventSeqDb store.EventSeq) *getDetailsUseCase {
	return &getDetailsUseCase{
		client: c,

		accountEraSeqDb: accountEraSeqDb,
		eventSeqDb:      eventSeqDb,
	}
}

func (uc *getDetailsUseCase) Execute(address string) (*DetailsView, error) {
	identity, err := uc.client.Account.GetIdentity(address)
	if err != nil {
		return nil, err
	}

	eraLimit := int64(1)
	accountEraSeqs, err := uc.accountEraSeqDb.FindLastByStashAccount(address, eraLimit)
	if err != nil {
		return nil, err
	}

	balanceTransfers, err := uc.eventSeqDb.FindBalanceTransfers(address)
	if err != nil {
		return nil, err
	}

	balanceDeposits, err := uc.eventSeqDb.FindBalanceDeposits(address)
	if err != nil {
		return nil, err
	}

	bonded, err := uc.eventSeqDb.FindBonded(address)
	if err != nil {
		return nil, err
	}

	unbonded, err := uc.eventSeqDb.FindUnbonded(address)
	if err != nil {
		return nil, err
	}

	withdrawn, err := uc.eventSeqDb.FindWithdrawn(address)
	if err != nil {
		return nil, err
	}

	return ToDetailsView(address, identity.GetIdentity(), accountEraSeqs, balanceTransfers, balanceDeposits, bonded, unbonded, withdrawn)
}
