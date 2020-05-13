package blockseqmapper

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

func Test_BlockSeqMapper(t *testing.T) {
	chainId := "chain123"
	model := &shared.Model{}
	sequence := &shared.Sequence{
		ChainId: chainId,
		Height:  types.Height(10),
		Time:    *types.NewTimeFromTime(time.Now()),
	}
	rep := report.Model{
		Model:        &shared.Model{},
		StartHeight:  types.Height(1),
		EndHeight:    types.Height(10),
	}
	blockFixture := fixtures.Load("block.json")
	validatorsFixture := fixtures.Load("validators.json")

	t.Run("ToSequence() fails unmarshal data when both block and validators syncables are bad", func(t *testing.T) {
		bs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   types.SyncableTypeBlock,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(`{"test": 0}`)},
		}

		vs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   syncable.ValidatorsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(`{"test": 0}`)},
		}

		_, err := ToSequence(bs, vs)
		if err == nil {
			t.Error("data unmarshaling should fail")
		}
	})

	t.Run("ToSequence() fails unmarshal data when both block syncable is bad", func(t *testing.T) {
		bs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   types.SyncableTypeBlock,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(`{"test": 0}`)},
		}

		vs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   syncable.ValidatorsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(validatorsFixture)},
		}

		_, err := ToSequence(bs, vs)
		if err == nil {
			t.Error("data unmarshaling should fail")
		}
	})

	t.Run("ToSequence() fails unmarshal data when validators syncable is bad", func(t *testing.T) {
		bs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   types.SyncableTypeBlock,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(blockFixture)},
		}

		vs := syncable.Model{
			Model:    model,
			Sequence: sequence,

			Type:   syncable.ValidatorsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(`{"test": 0}`)},
		}

		_, err := ToSequence(bs, vs)
		if err == nil {
			t.Error("data unmarshaling should fail")
		}
	})

	t.Run("ToSequence() succeeds unmarshal data", func(t *testing.T) {
		seq := &shared.Sequence{
			ChainId: chainId,
			Height:  types.Height(10),
			Time:    *types.NewTimeFromTime(time.Now()),
		}
		bs := syncable.Model{
			Model:       model,
			Sequence: seq,

			Type:   types.SyncableTypeBlock,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(blockFixture)},
		}

		vs := syncable.Model{
			Model:       model,
			Sequence: seq,

			Type:   syncable.ValidatorsType,
			Report: rep,
			Data:   types.Jsonb{RawMessage: json.RawMessage(validatorsFixture)},
		}

		blockSeq, err := ToSequence(bs, vs)
		if err != nil {
			t.Error("data unmarshaling should succeed")
			return
		}

		if blockSeq.Hash != "DCFA79417A55AAF9083AF654D4495EB91C5846B2B684770758D68073D92F8027" {
			t.Errorf("block hash expected: %s, got: %s", "DCFA79417A55AAF9083AF654D4495EB91C5846B2B684770758D68073D92F8027", blockSeq.Hash)
		}

		if blockSeq.ProposerEntityUID != "7hAEASOPAKuePwArPtGlcg3nnaipRdgqb0hESkA7Jso=" {
			t.Errorf("block proposer address expected: %s, got: %s", "7hAEASOPAKuePwArPtGlcg3nnaipRdgqb0hESkA7Jso=", blockSeq.ProposerEntityUID)
		}
	})
}
