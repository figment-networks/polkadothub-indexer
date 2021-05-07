package apr

import (
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type getAprByAddressUseCase struct {
	accountDb   store.AccountEraSeq
	rewardDb    store.Rewards
	syncablesDb store.Syncables
}

func NewGetAprByAddressUseCase(accountDb store.AccountEraSeq, rewardDb store.Rewards, syncablesDb store.Syncables) *getAprByAddressUseCase {
	return &getAprByAddressUseCase{
		accountDb:   accountDb,
		rewardDb:    rewardDb,
		syncablesDb: syncablesDb,
	}
}

func (uc *getAprByAddressUseCase) Execute(stash string, start, end types.Time) (res []DailyApr, err error) {
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return res, err
	}
	if mostRecentSynced.Time.Before(end.Time) {
		end = *types.NewTimeFromTime(mostRecentSynced.Time.Time)
	}

	summaries, err := uc.rewardDb.GetAllByTime(stash, start, end)
	if err != nil {
		return res, err
	}

	accountSeqs, err := uc.accountDb.GetAllByTime(stash, start, end)
	if err != nil {
		return res, err
	}

	return toAPRView(summaries, accountSeqs)
}
