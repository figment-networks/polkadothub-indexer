package model

// swagger:response BlockSummary
type BlockSummary struct {
	*Model
	*Summary

	Count        int64   `json:"count"`
	BlockTimeAvg float64 `json:"block_time_avg"`
}

func (BlockSummary) TableName() string {
	return "block_summary"
}
