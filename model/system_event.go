package model

import "github.com/figment-networks/polkadothub-indexer/types"

const (
	SystemEventActiveBalanceChange1 SystemEventKind = "active_balance_change_1"
	SystemEventActiveBalanceChange2 SystemEventKind = "active_balance_change_2"
	SystemEventActiveBalanceChange3 SystemEventKind = "active_balance_change_3"
	SystemEventCommissionChange1    SystemEventKind = "commission_change_1"
	SystemEventCommissionChange2    SystemEventKind = "commission_change_2"
	SystemEventCommissionChange3    SystemEventKind = "commission_change_3"
	SystemEventJoinedSet            SystemEventKind = "joined_set"
	SystemEventLeftSet              SystemEventKind = "left_set"
	SystemEventMissedNConsecutive   SystemEventKind = "missed_n_consecutive"
	SystemEventDelegationLeft       SystemEventKind = "delegation_left"
	SystemEventDelegationJoined     SystemEventKind = "delegation_joined"
)

type SystemEventKind string

func (o SystemEventKind) String() string {
	return string(o)
}

type SystemEvent struct {
	*Model

	Height int64           `json:"height"`
	Time   types.Time      `json:"time"`
	Actor  string          `json:"actor"`
	Kind   SystemEventKind `json:"kind"`
	Data   types.Jsonb     `json:"data"`
}

// DelegationChangeData is data format for delegation system events
type DelegationChangeData struct {
	StashAccounts []string `json:"stash_accounts"`
}

// PercentChangeData is data format for change system events
type PercentChangeData struct {
	Before int64   `json:"before"`
	After  int64   `json:"after"`
	Change float64 `json:"change"`
}

// MissedNofMData is data format for missedNofM system events
type MissedNofMData struct {
	Missed           int64 `json:"missed"`
	Threshold        int64 `json:"threshold"`
	MaxTotalSessions int64 `json:"max_total_sessions"`
}

// MissedNConsecutive is data format for missedNofM system events
type MissedNConsecutive struct {
	Missed    int64 `json:"missed"`
	Threshold int64 `json:"threshold"`
}
