package store

import "github.com/figment-networks/polkadothub-indexer/types"

type ActivityPeriodRow struct {
	Period int64
	Min    types.Time
	Max    types.Time
}
