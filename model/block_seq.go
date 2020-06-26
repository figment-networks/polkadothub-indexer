package model

import "github.com/figment-networks/polkadothub-indexer/types"

type BlockSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	// Indexed data
	ExtrinsicsCount         int64  `json:"extrinsics_count"`
	UnsignedExtrinsicsCount int64  `json:"unsigned_extrinsics_count"`
	SignedExtrinsicsCount   int64  `json:"signed_extrinsics_count"`
}

func (BlockSeq) TableName() string {
	return "block_sequences"
}

func (b *BlockSeq) Valid() bool {
	return b.Sequence.Valid()
}

func (b *BlockSeq) Equal(m BlockSeq) bool {
	return b.Sequence.Equal(*m.Sequence)
}

func (b *BlockSeq) Update(m BlockSeq) {
	b.ExtrinsicsCount = m.ExtrinsicsCount
	b.UnsignedExtrinsicsCount = m.UnsignedExtrinsicsCount
	b.SignedExtrinsicsCount = m.SignedExtrinsicsCount
}
