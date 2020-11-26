package model

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/lib/pq"
)

type ValidatorEraSeq struct {
	ID types.ID `json:"id"`

	*EraSequence

	StashAccount      string         `json:"stash_account"`
	ControllerAccount string         `json:"controller_account"`
	SessionAccounts   pq.StringArray `json:"session_accounts"`
	Index             int64          `json:"index"`
	TotalStake        types.Quantity `json:"total_stake"`
	OwnStake          types.Quantity `json:"own_stake"`
	StakersStake      types.Quantity `json:"stakers_stake"`
	RewardPoints      int64          `json:"reward_points"`
	Commission        int64          `json:"commission"`
	StakersCount      int            `json:"stakers_count"`
}

func (ValidatorEraSeq) TableName() string {
	return "validator_era_sequences"
}

func (s *ValidatorEraSeq) Valid() bool {
	return s.EraSequence.Valid() &&
		s.StashAccount != "" &&
		s.Index >= 0
}

func (s *ValidatorEraSeq) Equal(m ValidatorEraSeq) bool {
	return s.Index == m.Index &&
		s.StashAccount == m.StashAccount
}
