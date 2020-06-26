package stakingseq

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Model struct {
	*shared.Model
	*shared.HeightSequence

	TotalStake        types.Quantity `json:"total_stake"`
	TotalRewardPayout types.Quantity `json:"total_reward_payout"`
	TotalRewardPoints types.Quantity `json:"total_reward_points"`
}

// - Methods
func (Model) TableName() string {
	return "staking_sequences"
}

func (s *Model) ValidOwn() bool {
	return true
}

func (s *Model) EqualOwn(m Model) bool {
	return true
}

func (s *Model) Valid() bool {
	return s.Model.Valid() &&
		s.HeightSequence.Valid() &&
		s.ValidOwn()
}

func (s *Model) Equal(m Model) bool {
	return s.Model != nil &&
		m.Model != nil &&
		s.Model.Equal(*m.Model) &&
		s.EqualOwn(m)
}