package shared

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type HeightSequence struct {
	ChainUid       string       `json:"chain_uid"`
	SpecVersionUid string       `json:"spec_version_uid"`
	Height         types.Height `json:"height"`
	Time           *types.Time   `json:"time"`
}

func (s *HeightSequence) Valid() bool {
	return s.ChainUid != "" &&
		s.Height.Valid() &&
		!s.Time.IsZero()
}

func (s *HeightSequence) Equal(m HeightSequence) bool {
	return s.ChainUid == m.ChainUid &&
		s.Height.Equal(m.Height) &&
		s.Time.Equal(*m.Time)
}
