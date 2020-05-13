package extrinsicseq

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
)

type Model struct {
	*shared.Model
	*shared.HeightSequence

	Index     int64  `json:"index"`
	Signature string `json:"signature" `
	Signer    string `json:"signer"`
	Nonce     int64  `json:"nonce"`
	Method    string `json:"call"`
	Section   string `json:"module_id"`
	Args      string `json:"call"`
	IsSigned  bool   `json:"is_signed"`
}

// - Methods
func (Model) TableName() string {
	return "extrinsic_sequences"
}

func (ts *Model) ValidOwn() bool {
	return true
}

func (ts *Model) EqualOwn(m Model) bool {
	return ts.Index == m.Index
}

func (ts *Model) Valid() bool {
	return ts.HeightSequence.Valid() &&
		ts.ValidOwn()
}

func (ts *Model) Equal(m Model) bool {
	return ts.Model != nil &&
		m.Model != nil &&
		ts.Model.Equal(*m.Model) &&
		ts.HeightSequence.Equal(*m.HeightSequence) &&
		ts.EqualOwn(m)
}
