package blockseq

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
)

type Model struct {
	*shared.Model
	*shared.HeightSequence

	ParentHash              string     `json:"parent_hash"`
	StateRoot               string     `json:"state_root"`
	ExtrinsicsRoot          string     `json:"extrinsics_root" `
	ExtrinsicsCount         int64      `json:"extrinsics_count"`
	UnsignedExtrinsicsCount int64      `json:"unsigned_extrinsics_count"`
	SignedExtrinsicsCount   int64      `json:"signed_extrinsics_count"`
}

// - METHODS
func (Model) TableName() string {
	return "block_sequences"
}

func (b *Model) ValidOwn() bool {
	return !b.Time.IsZero()
}

func (b *Model) EqualOwn(m Model) bool {
	return b.ParentHash == m.ParentHash &&
		b.StateRoot == m.StateRoot &&
		b.ExtrinsicsRoot == m.ExtrinsicsRoot
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
