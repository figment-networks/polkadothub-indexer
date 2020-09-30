package model

import "github.com/figment-networks/polkadothub-indexer/types"

type ValidatorSessionSeqSummary struct {
	StashAccount string     `json:"stash_account"`
	TimeBucket   types.Time `json:"time_bucket"`
	UptimeAvg    float64    `json:"uptime_avg"`
	UptimeMin    int64      `json:"uptime_min"`
	UptimeMax    int64      `json:"uptime_max"`
}
