package account

import (
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/store"
)

type getRewardsUseCase struct {
	eventSeqDb  store.EventSeq
	syncablesDb store.Syncables
}

func NewGetRewardsUseCase(eventSeqDb store.EventSeq, syncablesDb store.Syncables) *getRewardsUseCase {
	return &getRewardsUseCase{
		eventSeqDb:  eventSeqDb,
		syncablesDb: syncablesDb,
	}
}

func (uc *getRewardsUseCase) Execute(address string, start, end time.Time) (RewardsView, error) {
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return RewardsView{}, err
	}
	if mostRecentSynced.Time.Time.Before(end) {
		return RewardsView{}, fmt.Errorf("block for time %v is not indexed yet; latest block time: %v", end.UTC(), mostRecentSynced.Time.UTC())
	}

	if end.IsZero() {
		end = mostRecentSynced.Time.Time
	}

	events, err := uc.eventSeqDb.FindRewardsForTimePeriod(address, start, end)
	if err != nil {
		return RewardsView{}, nil
	}

	return toRewardsView(events, address, start, end)
}
