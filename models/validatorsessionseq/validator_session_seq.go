package validatorsessionseq

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
)

type Model struct {
	*shared.Model
	*shared.SessionSequence

	StashAccount string `json:"stash_account"`
	Online       bool   `json:"online"`
}

// - Methods
func (Model) TableName() string {
	return "validator_session_sequences"
}

func (s *Model) ValidOwn() bool {
	return s.StashAccount != ""
}

func (s *Model) EqualOwn(m Model) bool {
	return s.StashAccount == m.StashAccount
}

func (s *Model) Valid() bool {
	return s.Model.Valid() &&
		s.SessionSequence.Valid() &&
		s.ValidOwn()
}

func (s *Model) Equal(m Model) bool {
	return s.Model != nil &&
		m.Model != nil &&
		s.Model.Equal(*m.Model) &&
		s.EqualOwn(m)
}
