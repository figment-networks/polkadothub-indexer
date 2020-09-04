package common

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Delegation struct {
	Era            int64          `json:"era"`
	ValidatorStash string         `json:"validator_stash"`
	Amount         types.Quantity `json:"amount"`
}

func ToDelegations(accountEraSeqs []model.AccountEraSeq) []*Delegation {
	var delegations []*Delegation
	for _, accountEraSeq := range accountEraSeqs {
		delegations = append(delegations, &Delegation{
			Era:            accountEraSeq.Era,
			ValidatorStash: accountEraSeq.ValidatorStashAccount,
			Amount:         accountEraSeq.Stake,
		})
	}
	return delegations
}
