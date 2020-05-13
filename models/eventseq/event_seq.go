package eventseq

import "github.com/figment-networks/polkadothub-indexer/models/shared"

type Model struct {
	*shared.Model
	*shared.HeightSequence

	EventIdx     int64  `json:"event_idx"`
	ExtrinsicIdx int64  `json:"extrinsic_idx"`
	Type         string `json:"type"`
	ModuleId     string `json:"module_id"`
	EventId      string `json:"event_id"`
	System       int64  `json:"system"`
	Module       int64  `json:"module"`
	Phase        int64  `json:"module"`
	Attributes   string `json:"attributes"`
}

// - METHODS
func (Model) TableName() string {
	return "event_sequences"
}

func (b *Model) ValidOwn() bool {
	return !b.Time.IsZero()
}

func (b *Model) EqualOwn(m Model) bool {
	return b.EventIdx == m.EventIdx &&
		b.System == m.System &&
		b.Module == m.Module
}

func (b *Model) Valid() bool {
	return b.Model.Valid() &&
		b.HeightSequence.Valid() &&
		b.ValidOwn()
}

func (b *Model) Equal(m Model) bool {
	return b.Model != nil &&
		m.Model != nil &&
		b.Model.Equal(*m.Model) &&
		b.HeightSequence.Equal(*m.HeightSequence) &&
		b.EqualOwn(m)
}
