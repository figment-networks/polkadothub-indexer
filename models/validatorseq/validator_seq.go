package validatorseq

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Model struct {
	*shared.Model
	*shared.HeightSequence

	Index             int64          `json:"index"`
	StashAccount      string         `json:"stash_account"`
	ControllerAccount string         `json:"controller_account"`
	SessionAccount    string         `json:"session_account"`
	TotalStake        types.Quantity `json:"total_stake"`
	OwnStake          types.Quantity `json:"own_stake"`
	RewardPoints      int64          `json:"reward_points"`
	Commission        int64          `json:"commission"`
	//Unlocking           string `json:"unlocking"`
	NominatorsCount int64 `json:"nominators_count"`
	Online          *bool  `json:"online"`
	//UnstakeThreshold    int64  `json:"unstake_threshold"`
}

// - Methods
func (Model) TableName() string {
	return "validator_sequences"
}

func (s *Model) ValidOwn() bool {
	return s.Index >= 0
}

func (s *Model) EqualOwn(m Model) bool {
	return s.Index == m.Index
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
