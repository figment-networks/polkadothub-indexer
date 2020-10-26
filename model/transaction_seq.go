package model

import "github.com/figment-networks/polkadothub-indexer/types"

type TransactionSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	// Indexed data
	Index   int64  `json:"index"`
	Hash    string `json:"hash"`
	Method  string `json:"method"`
	Section string `json:"section"`
}

func (TransactionSeq) TableName() string {
	return "transaction_sequences"
}

func (b *TransactionSeq) Valid() bool {
	return b.Sequence.Valid()
}

func (b *TransactionSeq) Update(m TransactionSeq) {
	b.Index = m.Index
	b.Hash = m.Hash
	b.Method = m.Method
	b.Section = m.Section
}
