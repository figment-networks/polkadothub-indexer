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

func (t *TransactionSeq) Valid() bool {
	if t.Hash == "" || t.Method == "" || t.Section == "" || !t.Sequence.Valid() {
		return false
	}
	return true
}

func (t *TransactionSeq) Update(m TransactionSeq) {
	t.Index = m.Index
	t.Hash = m.Hash
	t.Method = m.Method
	t.Section = m.Section
}
