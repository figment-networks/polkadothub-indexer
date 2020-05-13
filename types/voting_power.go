package types

type VotingPower int64

func (vp VotingPower) Valid() bool {
	return vp >= 0
}

func (vp VotingPower) Equal(o VotingPower) bool {
	return vp == o
}
