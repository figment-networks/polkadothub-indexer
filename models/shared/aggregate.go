package shared

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Aggregate struct {
	ChainUid        string       `json:"chain_uid"`
	SpecVersionUid  string       `json:"spec_version_uid"`
	StartedAtHeight types.Height `json:"started_at_height"`
	StartedAt       types.Time   `json:"started_at"`
	RecentAtHeight  types.Height `json:"recent_at_height"`
	RecentAt        types.Time   `json:"recent_at"`
}

func (a *Aggregate) Valid() bool {
	return a.StartedAtHeight.Valid() &&
		!a.StartedAt.IsZero()
}

func (a *Aggregate) Equal(m Aggregate) bool {
	return a.StartedAtHeight.Equal(m.StartedAtHeight) &&
		a.StartedAt.Equal(m.StartedAt)
}
