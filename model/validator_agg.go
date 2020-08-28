package model

type ValidatorAgg struct {
	*Model
	*Aggregate

	StashAccount            string `json:"stash_account"`
	RecentAsValidatorHeight int64  `json:"recent_as_validator_height"`
	AccumulatedUptime       int64  `json:"accumulated_uptime"`
	AccumulatedUptimeCount  int64  `json:"accumulated_uptime_count"`
}

// - Methods
func (ValidatorAgg) TableName() string {
	return "validator_aggregates"
}

func (s *ValidatorAgg) Valid() bool {
	return s.Aggregate.Valid() &&
		s.StashAccount != ""
}

func (s *ValidatorAgg) Equal(m ValidatorAgg) bool {
	return s.StashAccount == m.StashAccount
}

func (s *ValidatorAgg) Update(u *ValidatorAgg) {
	s.Aggregate.RecentAtHeight = u.Aggregate.RecentAtHeight
	s.Aggregate.RecentAt = u.Aggregate.RecentAt
}
