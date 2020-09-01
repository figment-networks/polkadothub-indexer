package model

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type AccountEraSeq struct {
	ID types.ID `json:"id"`

	*EraSequence

	StashAccount               string         `json:"stash_account"`
	ControllerAccount          string         `json:"controller_account"`
	ValidatorStashAccount      string         `json:"validator_stash_account"`
	ValidatorControllerAccount string         `json:"validator_controller_account"`
	Stake                      types.Quantity `json:"own_stake"`
}

func (AccountEraSeq) TableName() string {
	return "account_era_sequences"
}

func (s *AccountEraSeq) Valid() bool {
	return s.EraSequence.Valid() &&
		s.ValidatorStashAccount != "" &&
		s.ControllerAccount != ""
}

func (s *AccountEraSeq) Equal(m AccountEraSeq) bool {
	return s.ValidatorStashAccount == m.ValidatorStashAccount &&
		s.ControllerAccount == m.ControllerAccount
}

func (b *AccountEraSeq) Update(m AccountEraSeq) {
	b.ValidatorControllerAccount = m.ValidatorControllerAccount
	b.ControllerAccount = m.ControllerAccount
	b.Stake = m.Stake
}
