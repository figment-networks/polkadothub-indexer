package validator

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/common"
	"github.com/lib/pq"
)

// swagger:response ValidatorAggregatesView
type AggListView struct {
	Items []model.ValidatorAgg `json:"items"`
}

func ToAggListView(ms []model.ValidatorAgg) *AggListView {
	return &AggListView{
		Items: ms,
	}
}

// swagger:response ValidatorAggregateView
type AggDetailsView struct {
	*model.Model
	*model.Aggregate

	StashAccount            string  `json:"stash_account"`
	RecentAsValidatorHeight int64   `json:"recent_as_validator_height"`
	DisplayName             string  `json:"display_name"`
	Uptime                  float64 `json:"uptime"`

	LastSessionSequences []model.ValidatorSessionSeq `json:"last_session_sequences"`
	LastEraSequences     []model.ValidatorEraSeq     `json:"last_era_sequences"`
	LastDelegations      []*common.Delegation        `json:"delegations"`
}

func ToAggDetailsView(m *model.ValidatorAgg, sessionSequences []model.ValidatorSessionSeq, eraSequences []model.ValidatorEraSeq, accountEraSequences []model.AccountEraSeq) *AggDetailsView {
	return &AggDetailsView{
		Model:     m.Model,
		Aggregate: m.Aggregate,

		StashAccount:            m.StashAccount,
		RecentAsValidatorHeight: m.RecentAsValidatorHeight,
		DisplayName:             m.DisplayName,
		Uptime:                  float64(m.AccumulatedUptime) / float64(m.AccumulatedUptimeCount),

		LastSessionSequences: sessionSequences,
		LastEraSequences:     eraSequences,
		LastDelegations:      common.ToDelegations(accountEraSequences),
	}
}

type SessionSeqListItem struct {
	*model.SessionSequence

	DisplayName  string `json:"display_name"`
	StashAccount string `json:"stash_account"`
	Online       bool   `json:"online"`
}

type EraSeqListItem struct {
	*model.EraSequence

	StashAccount      string         `json:"stash_account"`
	ControllerAccount string         `json:"controller_account"`
	SessionAccounts   pq.StringArray `json:"session_accounts"`
	Index             int64          `json:"index"`
	TotalStake        types.Quantity `json:"total_stake"`
	OwnStake          types.Quantity `json:"own_stake"`
	StakersStake      types.Quantity `json:"stakers_stake"`
	RewardPoints      int64          `json:"reward_points"`
	Commission        int64          `json:"commission"`
	StakersCount      int            `json:"stakers_count"`
}

// swagger:response ValidatorsView
type SeqListView struct {
	SessionItems []SessionSeqListItem `json:"session_items"`
	EraItems     []EraSeqListItem     `json:"era_items"`
}

func ToSeqListView(validatorSessionSeqs []model.ValidatorSessionSeq, validatorEraSeqs []model.ValidatorEraSeq) SeqListView {
	var sessionItems []SessionSeqListItem
	for _, m := range validatorSessionSeqs {
		item := SessionSeqListItem{
			SessionSequence: m.SessionSequence,

			StashAccount: m.StashAccount,
			Online:       m.Online,
		}

		sessionItems = append(sessionItems, item)
	}

	var eraItems []EraSeqListItem
	for _, m := range validatorEraSeqs {
		item := EraSeqListItem{
			EraSequence: m.EraSequence,

			StashAccount:      m.StashAccount,
			ControllerAccount: m.ControllerAccount,
			SessionAccounts:   m.SessionAccounts,
			Index:             m.Index,
			TotalStake:        m.TotalStake,
			OwnStake:          m.OwnStake,
			StakersStake:      m.StakersStake,
			RewardPoints:      m.RewardPoints,
			Commission:        m.Commission,
			StakersCount:      m.StakersCount,
		}

		eraItems = append(eraItems, item)
	}

	return SeqListView{
		SessionItems: sessionItems,
		EraItems:     eraItems,
	}
}

type syncableInfo struct {
	Time types.Time `json:"time"`
}

// swagger:response ValidatorsSummariesView
type summaryListView struct {
	Items       []store.ValidatorSummaryRow `json:"items"`
	LastSession *syncableInfo               `json:"last_indexed_session,omitempty"`
	LastEra     *syncableInfo               `json:"last_indexed_era,omitempty"`
}

func toSummaryListView(summaries []store.ValidatorSummaryRow, lastIndexedSession, lastIndexedEra *model.Syncable) summaryListView {
	view := summaryListView{
		Items: summaries,
	}

	if lastIndexedSession != nil {
		view.LastSession = &syncableInfo{Time: lastIndexedSession.Time}
	}

	if lastIndexedEra != nil {
		view.LastEra = &syncableInfo{Time: lastIndexedEra.Time}
	}

	return view
}
