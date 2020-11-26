package model

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type ValidatorSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	StashAccount  string         `json:"stash_account"`
	ActiveBalance types.Quantity `json:"active_balance"`
}

func (ValidatorSeq) TableName() string {
	return "validator_sequences"
}

func (vs *ValidatorSeq) Valid() bool {
	return vs.Sequence.Valid() &&
		vs.StashAccount != "" &&
		vs.ActiveBalance.Valid()
}

func (vs *ValidatorSeq) Equal(m ValidatorSeq) bool {
	return vs.Sequence.Equal(*m.Sequence) &&
		vs.StashAccount == m.StashAccount
}
