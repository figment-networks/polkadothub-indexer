package account

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getDetailsUseCase struct {
	client *client.Client

	accountEraSeqDb store.AccountEraSeq
	eventSeqDb      store.EventSeq
	syncablesDb     store.Syncables
}

func NewGetDetailsUseCase(c *client.Client, accountEraSeqDb store.AccountEraSeq, eventSeqDb store.EventSeq, syncablesDb store.Syncables) *getDetailsUseCase {
	return &getDetailsUseCase{
		client: c,

		accountEraSeqDb: accountEraSeqDb,
		eventSeqDb:      eventSeqDb,
		syncablesDb:     syncablesDb,
	}
}

func (uc *getDetailsUseCase) Execute(address string) (DetailsView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return DetailsView{}, err
	}
	lastH := mostRecentSynced.Height

	account, err := uc.client.Account.GetByHeight(address, lastH)
	if err != nil {
		return DetailsView{}, err
	}

	identity, err := uc.client.Account.GetIdentity(address)
	if err != nil {
		return DetailsView{}, err
	}

	accountEraSeqs, err := uc.accountEraSeqDb.FindLastByStashAccount(address)
	if err != nil {
		return DetailsView{}, err
	}

	balanceTransfers, err := uc.eventSeqDb.FindBalanceTransfers(address)
	if err != nil {
		return DetailsView{}, err
	}

	balanceDeposits, err := uc.eventSeqDb.FindBalanceDeposits(address)
	if err != nil {
		return DetailsView{}, err
	}

	bonded, err := uc.eventSeqDb.FindBonded(address)
	if err != nil {
		return DetailsView{}, err
	}

	unbonded, err := uc.eventSeqDb.FindUnbonded(address)
	if err != nil {
		return DetailsView{}, err
	}

	withdrawn, err := uc.eventSeqDb.FindWithdrawn(address)
	if err != nil {
		return DetailsView{}, err
	}

	return ToDetailsView(address, identity.GetIdentity(), account.GetAccount(), accountEraSeqs, balanceTransfers, balanceDeposits, bonded, unbonded, withdrawn)
}
