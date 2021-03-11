package model

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/types"
)

const (
	SyncableStatusRunning SyncableStatus = iota + 1
	SyncableStatusCompleted
)

type SyncableStatus int

type Syncable struct {
	*Model

	Height        int64      `json:"height"`
	Time          types.Time `json:"time"`
	SpecVersion   string     `json:"spec_version"`
	ChainUID      string     `json:"chain_uid"`
	Session       int64      `json:"session"`
	Era           int64      `json:"era"`
	LastInSession bool       `json:"last_in_session"`
	LastInEra     bool       `json:"last_in_era"`

	IndexVersion int64          `json:"index_version"`
	Status       SyncableStatus `json:"status"`
	ReportID     types.ID       `json:"report_id"`
	StartedAt    types.Time     `json:"started_at"`
	ProcessedAt  *types.Time    `json:"processed_at"`
	Duration     time.Duration  `json:"duration"`
}

// - Methods
func (Syncable) TableName() string {
	return "syncables"
}

func (s *Syncable) Valid() bool {
	return s.Height >= 0
}

func (s *Syncable) Equal(m Syncable) bool {
	return s.Height == m.Height
}

func (s *Syncable) SetStatus(newStatus SyncableStatus) {
	s.Status = newStatus
}

func (s *Syncable) MarkProcessed(indexVersion int64) {
	t := types.NewTimeFromTime(time.Now())
	duration := time.Since(s.StartedAt.Time)

	s.Status = SyncableStatusCompleted
	s.Duration = duration
	s.ProcessedAt = t
	s.IndexVersion = indexVersion
}
