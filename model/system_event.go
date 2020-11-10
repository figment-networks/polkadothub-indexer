package model

import "github.com/figment-networks/polkadothub-indexer/types"

const (
	SystemEventActiveBalanceChange1 SystemEventKind = "active_balance_change_1"
	SystemEventActiveBalanceChange2 SystemEventKind = "active_balance_change_2"
	SystemEventActiveBalanceChange3 SystemEventKind = "active_balance_change_3"
	SystemEventCommissionChange1    SystemEventKind = "commission_change_1"
	SystemEventCommissionChange2    SystemEventKind = "commission_change_2"
	SystemEventCommissionChange3    SystemEventKind = "commission_change_3"
	SystemEventJoinedActiveSet      SystemEventKind = "joined_active_set"
	SystemEventJoinedWaitingSet     SystemEventKind = "joined_waiting_set"
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

// MissedNConsecutive is data format for missed_n_consecutive system events
type MissedNConsecutive struct {
	Missed    int64 `json:"missed"`
	Threshold int64 `json:"threshold"`
}

func (o SystemEvent) Update(m SystemEvent) {
	o.Height = m.Height
	o.Time = m.Time
	o.Actor = m.Actor
	o.Kind = m.Kind
	o.Data = m.Data
}
