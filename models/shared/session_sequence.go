package shared

type SessionSequence struct {
	ChainUid        string       `json:"chain_uid"`
	SpecVersionUid  string       `json:"spec_version_uid"`
	Session int64  `json:"session"`
}

func (s *SessionSequence) Valid() bool {
	return s.Session >= 0
}

func (s *SessionSequence) Equal(m SessionSequence) bool {
	return s.Session == m.Session
}
