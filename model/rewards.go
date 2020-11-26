package model

const (
	UnclaimedCommission RewardKind = "unclaimed_commission"
	UnclaimedReward     RewardKind = "unclaimed_reward"
)

type RewardKind string

func (o RewardKind) String() string {
	return string(o)
}

type Reward struct {
	*Model

	Era                   int64      `json:"era"`
	StashAccount          string     `json:"stash_account"`
	ValidatorStashAccount string     `json:"validator_stash_account"`
	Amount                string     `json:"amount"`
	Kind                  RewardKind `json:"kind"`
}
