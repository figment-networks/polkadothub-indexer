package model

import "github.com/figment-networks/polkadothub-indexer/types"

type ValidatorSessionSeq struct {
	ID types.ID `json:"id"`

	*SessionSequence

	StashAccount string `json:"stash_account"`
	Online       bool   `json:"online"`
}

func (ValidatorSessionSeq) TableName() string {
	return "validator_session_sequences"
}

func (s *ValidatorSessionSeq) Valid() bool {
	return s.SessionSequence.Valid()
}

func (s *ValidatorSessionSeq) Equal(m ValidatorSessionSeq) bool {
	return s.SessionSequence.Equal(*m.SessionSequence)
}

func (b *ValidatorSessionSeq) Update(m ValidatorSessionSeq) {
	b.Online = m.Online
}
