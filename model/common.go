package model

import "github.com/figment-networks/polkadothub-indexer/types"

// Model is used for general table
type Model struct {
	ID        types.ID   `json:"id"`
	CreatedAt types.Time `json:"created_at"`
	UpdatedAt types.Time `json:"updated_at"`
}

func (e *Model) Valid() bool {
	return true
}

func (e *Model) Equal(m Model) bool {
	return e.ID.Equal(m.ID)
}

// Sequence is used for sequence tables
type Sequence struct {
	Height int64      `json:"height"`
	Time   types.Time `json:"time"`
}

func (s *Sequence) Valid() bool {
	return s.Height >= 0 &&
		!s.Time.IsZero()
}

func (s *Sequence) Equal(m Sequence) bool {
	return s.Height == m.Height &&
		s.Time.Equal(m.Time)
}

// SessionSequence is used for sequences recorded per session
type SessionSequence struct {
	Session     int64 `json:"session"`
	StartHeight int64 `json:"start_height"`
	EndHeight   int64 `json:"end_height"`
}

func (s *SessionSequence) Valid() bool {
	return s.Session >= 0 &&
		s.StartHeight >= 0 &&
		s.EndHeight >= 0
}

func (s *SessionSequence) Equal(m SessionSequence) bool {
	return s.Session == m.Session
}

// EraSequence is used for sequences recorded per era
type EraSequence struct {
	Era         int64 `json:"era"`
	StartHeight int64 `json:"start_height"`
	EndHeight   int64 `json:"end_height"`
}

func (s *EraSequence) Valid() bool {
	return s.Era >= 0 &&
		s.StartHeight >= 0 &&
		s.EndHeight >= 0
}

func (s *EraSequence) Equal(m EraSequence) bool {
	return s.Era == m.Era
}

// Aggregate is used for aggregate tables
type Aggregate struct {
	StartedAtHeight int64      `json:"started_at_height"`
	StartedAt       types.Time `json:"started_at"`
	RecentAtHeight  int64      `json:"recent_at_height"`
	RecentAt        types.Time `json:"recent_at"`
}

func (a *Aggregate) Valid() bool {
	return a.StartedAtHeight >= 0 &&
		!a.StartedAt.IsZero()
}

func (a *Aggregate) Equal(m Aggregate) bool {
	return a.StartedAtHeight == m.StartedAtHeight &&
		a.StartedAt.Equal(m.StartedAt)
}

// Summary is used for summary tables
type Summary struct {
	IndexVersion int64                 `json:"index_version"`
	TimeInterval types.SummaryInterval `json:"time_interval"`
	TimeBucket   types.Time            `json:"time_bucket"`
}

func (a *Summary) Valid() bool {
	return a.TimeInterval.Valid() &&
		!a.TimeBucket.IsZero()
}

func (a *Summary) Equal(m Summary) bool {
	return a.TimeInterval.Equal(m.TimeInterval) &&
		a.TimeBucket.Equal(m.TimeBucket)
}
