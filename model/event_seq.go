package model

import "github.com/figment-networks/polkadothub-indexer/types"

type EventSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	// Indexed data
	Index          int64       `json:"index"`
	ExtrinsicIndex int64       `json:"extrinsic_index"`
	Data           types.Jsonb `json:"data"`
	Phase          string      `json:"phase"`
	Method         string      `json:"method"`
	Section        string      `json:"section"`
}

type EventSeqWithTxHash struct {
	Height  int64       `json:"height"`
	Data    types.Jsonb `json:"data"`
	Method  string      `json:"method"`
	Section string      `json:"section"`
	TxHash  string      `json:"hash"`
}

func (EventSeq) TableName() string {
	return "event_sequences"
}

func (b *EventSeq) Valid() bool {
	return b.Sequence.Valid()
}

func (b *EventSeq) Equal(m EventSeq) bool {
	return b.Sequence.Equal(*m.Sequence)
}

func (b *EventSeq) Update(m EventSeq) {
	b.Index = m.Index
	b.ExtrinsicIndex = m.ExtrinsicIndex
	b.Data = m.Data
	b.Phase = m.Phase
	b.Method = m.Method
	b.Section = m.Section
}
