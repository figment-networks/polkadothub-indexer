package reward

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getForStashAccountUseCase struct {
	rewardDb store.Rewards
}

func NewGetForStashAccountUseCase(rewardDb store.Rewards) *getForStashAccountUseCase {
	return &getForStashAccountUseCase{
		rewardDb: rewardDb,
	}
}

// swagger:response RewardsForErasView
type RewardsForErasView []model.RewardEraSeq

func (uc *getForStashAccountUseCase) Execute(stash string, start, end int64, validatorStash string) ([]model.RewardEraSeq, error) {
	rewards, err := uc.rewardDb.GetAll(stash, validatorStash, start, end)
	if err != nil {
		return nil, err
	}

	return rewards, nil
}
