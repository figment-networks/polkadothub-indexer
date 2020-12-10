package model

const (
	RewardCommission RewardKind = "commission"
	RewardReward     RewardKind = "reward"
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
}
