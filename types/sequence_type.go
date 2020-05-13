package types

const (
	SequenceTypeHeight  = "height"
	SequenceTypeSession = "session"
	SequenceTypeEra     = "era"
)

var SequenceTypes = []SequenceType{SequenceTypeHeight, SequenceTypeSession, SequenceTypeEra}

type SequenceType string

func (t SequenceType) Valid() bool {
	return t == SequenceTypeHeight ||
		t == SequenceTypeSession ||
		t == SequenceTypeEra
}
