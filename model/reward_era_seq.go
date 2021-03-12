package model

const (
	RewardCommission RewardKind = "commission"
	RewardReward     RewardKind = "reward"
	// for historical data, it's not possible to dissect reward from commission
	RewardCommissionAndReward RewardKind = "commission_and_reward"
)

type RewardKind string

func (o RewardKind) String() string {
	return string(o)
}

type RewardEraSeq struct {
	*EraSequence

	StashAccount          string     `json:"stash_account"`
	ValidatorStashAccount string     `json:"validator_stash_account"`
	Amount                string     `json:"amount"`
	Kind                  RewardKind `json:"kind"`
	Claimed               bool       `json:"claimed"`
	TxHash                string     `json:"tx_hash"`
}

func (RewardEraSeq) TableName() string {
	return "reward_era_sequences"
}
