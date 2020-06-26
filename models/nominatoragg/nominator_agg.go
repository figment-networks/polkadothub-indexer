package nominatoragg

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Model struct {
	*shared.Model
	*shared.Aggregate

	StashAccount            string         `json:"stash_account"`
	RecentControllerAccount string         `json:"recent_controller_account"`
	RecentIndex             int64          `json:"recent_index"`
	RecentValidatorIndex    int64          `json:"recent_validator_index"`
	RecentStake             types.Quantity `json:"recent_stake"`
}

// - Methods
func (Model) TableName() string {
	return "nominator_aggregates"
}

func (s *Model) ValidOwn() bool {
	return s.StashAccount != ""
}

func (s *Model) EqualOwn(m Model) bool {
	return s.StashAccount == m.StashAccount
}

func (s *Model) Valid() bool {
	return s.Model.Valid() &&
		s.Aggregate.Valid() &&
		s.ValidOwn()
}

func (s *Model) Equal(m Model) bool {
	return s.Model != nil &&
		m.Model != nil &&
		s.Model.Equal(*m.Model) &&
		s.EqualOwn(m)
}

func (s *Model) UpdateAggAttrs(u *Model) {
	s.Aggregate.RecentAtHeight = u.Aggregate.RecentAtHeight
	s.Aggregate.RecentAt = u.Aggregate.RecentAt

	s.RecentControllerAccount = u.RecentControllerAccount
	s.RecentIndex = u.RecentIndex
	s.RecentValidatorIndex = u.RecentValidatorIndex
	s.RecentStake = u.RecentStake
}
