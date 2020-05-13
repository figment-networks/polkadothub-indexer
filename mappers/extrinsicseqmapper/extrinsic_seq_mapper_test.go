package extrinsicseqmapper

import (
	"encoding/json"
	"github.com/figment-networks/polkadothub-indexer/fixtures"
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"testing"
	"time"
)

func Test_StakingSeqMapper(t *testing.T) {
	chainId := "chain123"
	model := &shared.Model{}
	sequence := &shared.Sequence{
		ChainId: chainId,
		Height:  types.Height(10),
		Time:    *types.NewTimeFromTime(time.Now()),
	}
	rep := report.Model{
		Model:       &shared.Model{},
		StartHeight: types.Height(1),
		EndHeight:   types.Height(10),
	}
	transactionsFixture := fixtures.Load("transactions.json")

	t.Run("ToSequence()() fails unmarshal data", func(t *testing.T) {
		s := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   syncable.ExtrinsicsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(`{"test": 0}`)},
		}

		_, err := ToSequence(s)
		if err == nil {
			t.Error("data unmarshaling should fail")
		}
	})

	t.Run("ToSequence()() succeeds to unmarshal data", func(t *testing.T) {
		s := syncable.Model{
			Model: model,
			Sequence: sequence,

			Type:   syncable.ExtrinsicsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(transactionsFixture)},
		}

		transactionSeqs, err := ToSequence(s)
		if err != nil {
			t.Error("data unmarshaling should succeed", err)
		}

		if len(transactionSeqs) == 0 {
			t.Error("there should be transactions")
		}
	})
}


