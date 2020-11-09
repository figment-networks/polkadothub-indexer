package model

import "github.com/figment-networks/polkadothub-indexer/types"

type BlockSeqSummary struct {
	TimeBucket   types.Time `json:"time_bucket"`
	Count        int64      `json:"count"`
	BlockTimeAvg float64    `json:"block_time_avg"`
}
