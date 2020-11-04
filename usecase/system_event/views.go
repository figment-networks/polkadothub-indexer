package system_event

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type ListItem struct {
	*model.Model

	Height int64       `json:"height"`
	Time   types.Time  `json:"time"`
	Actor  string      `json:"actor"`
	Kind   string      `json:"kind"`
	Data   types.Jsonb `json:"data"`
}

type ListView struct {
	Items []ListItem `json:"items"`
}

func ToListView(events []model.SystemEvent) *ListView {
	items := make([]ListItem, len(events))
	for i, m := range events {
		items[i] = ListItem{
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
