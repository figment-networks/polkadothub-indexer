package shared

type EraSequence struct {
	ChainUid        string       `json:"chain_uid"`
	SpecVersionUid  string       `json:"spec_version_uid"`
	Era     int64  `json:"era"`
}

func (s *EraSequence) Valid() bool {
	return s.Era >= 0
}

func (s *EraSequence) Equal(m EraSequence) bool {
	return s.Era == m.Era
}
