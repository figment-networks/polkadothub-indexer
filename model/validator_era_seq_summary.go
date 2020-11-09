package model

import "github.com/figment-networks/polkadothub-indexer/types"

type ValidatorEraSeqSummary struct {
	StashAccount    string         `json:"stash_account"`
	TimeBucket      types.Time     `json:"time_bucket"`
	TotalStakeAvg   types.Quantity `json:"total_stake_avg"`
	TotalStakeMin   types.Quantity `json:"total_stake_min"`
	TotalStakeMax   types.Quantity `json:"total_stake_max"`
	OwnStakeAvg     types.Quantity `json:"own_stake_avg"`
	OwnStakeMin     types.Quantity `json:"own_stake_min"`
	OwnStakeMax     types.Quantity `json:"own_stake_max"`
	StakersStakeAvg types.Quantity `json:"stakers_stake_avg"`
	StakersStakeMin types.Quantity `json:"stakers_stake_min"`
	StakersStakeMax types.Quantity `json:"stakers_stake_max"`
	RewardPointsAvg float64        `json:"reward_points_avg"`
	RewardPointsMin int64          `json:"reward_points_min"`
	RewardPointsMax int64          `json:"reward_points_max"`
	CommissionAvg   float64        `json:"commission_avg"`
	CommissionMin   int64          `json:"commission_min"`
	CommissionMax   int64          `json:"commission_max"`
	StakersCountAvg float64        `json:"stakers_count_avg"`
	StakersCountMin int64          `json:"stakers_count_min"`
	StakersCountMax int64          `json:"stakers_count_max"`
}
