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
	SystemEventLeftActiveSet        SystemEventKind = "left_active_set"
	SystemEventMissedNConsecutive   SystemEventKind = "missed_n_consecutive"
	SystemEventMissedNofM           SystemEventKind = "missed_n_of_m"
)

type SystemEventKind string

func (o SystemEventKind) String() string {
	return string(o)
}

type SystemEvent struct {
	*Model

	Height int64           `json:"height"`
	Actor  string          `json:"actor"`
	Kind   SystemEventKind `json:"kind"`
	Data   types.Jsonb     `json:"data"`
}

func (o SystemEvent) Update(m SystemEvent) {
	o.Height = m.Height
	o.Actor = m.Actor
	o.Kind = m.Kind
	o.Data = m.Data
}
