package system_event

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type SystemEventItem struct {
	*model.Model

	// Height is block height from when event was triggered
	Height int64 `json:"height"`
	// Time is block time from when event was triggered
	Time types.Time `json:"time"`
	// Actor is an account stash
	Actor string `json:"actor"`
	// Kind is system event kind
	//
	// example: active_balance_change_1
	Kind string `json:"kind"`
	// Data contains event specific data for kind
	Data types.Jsonb `json:"data"`
}

// SystemEventsView is a list of system events
// swagger:response SystemEventsView
type ListView struct {
	Items []SystemEventItem `json:"items"`
}

func ToListView(events []model.SystemEvent) *ListView {
	items := make([]SystemEventItem, len(events))
	for i, m := range events {
		items[i] = SystemEventItem{
			Actor:  m.Actor,
			Height: m.Height,
			Time:   m.Time,
			Kind:   m.Kind.String(),
			Data:   m.Data,
		}
	}

	return &ListView{
		Items: items,
	}
}
