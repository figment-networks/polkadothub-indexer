package model

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type AccountEraSeq struct {
	ID types.ID `json:"id"`

	*EraSequence

	// Origin accounts
	StashAccount      string `json:"stash_account"`
	ControllerAccount string `json:"controller_account"`
	// Destination accounts
	ValidatorStashAccount      string         `json:"validator_stash_account"`
	ValidatorControllerAccount string         `json:"validator_controller_account"`
	Stake                      types.Quantity `json:"own_stake"`
}

func (AccountEraSeq) TableName() string {
	return "account_era_sequences"
}

func (s *AccountEraSeq) Valid() bool {
	return s.EraSequence.Valid() &&
		s.ValidatorStashAccount != ""
}

func (s *AccountEraSeq) Equal(m AccountEraSeq) bool {
	return s.ValidatorStashAccount == m.ValidatorStashAccount &&
		s.ValidatorControllerAccount == m.ValidatorControllerAccount
}
