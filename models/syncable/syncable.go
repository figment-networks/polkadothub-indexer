package syncable

import (
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
	"time"
)

type Model struct {
	*shared.Model

	SequenceType types.SequenceType
	SequenceId   types.SequenceId

	Type        types.SyncableType
	Report      report.Model `gorm:"foreignkey"`
	ReportID    *types.ID
	Data        types.Jsonb
	ProcessedAt *types.Time
}

// - Methods
func (Model) TableName() string {
	return "syncables"
}

func (s *Model) ValidOwn() bool {
	return s.Type.Valid()
}

func (s *Model) EqualOwn(m Model) bool {
	return s.Type == m.Type
}

func (s *Model) Valid() bool {
	return s.Model.Valid() &&
		s.ValidOwn()
}

func (s *Model) Equal(m Model) bool {
	return s.Model != nil &&
		m.Model != nil &&
		s.Model.Equal(*m.Model) &&
		s.EqualOwn(m)
}

func (s *Model) MarkProcessed(reportID types.ID) {
	t := types.NewTimeFromTime(time.Now())
	rid := reportID

	s.ProcessedAt = t
	s.ReportID = &rid
}
