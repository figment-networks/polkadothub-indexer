package shared

import (
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Model struct {
	ID        types.ID  `json:"id"`
	CreatedAt types.Time `json:"created_at"`
	UpdatedAt types.Time `json:"updated_at"`
}

//- Methods
func (e *Model) Valid() bool {
	return true
}

func (e *Model) Equal(m Model) bool {
	return e.ID.Equal(m.ID)
}
