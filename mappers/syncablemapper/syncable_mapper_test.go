package syncablemapper

import (
	"encoding/json"
	"github.com/figment-networks/polkadothub-indexer/fixtures"
	"github.com/figment-networks/polkadothub-indexer/types"
	"testing"
)

func Test_SyncableMapper_UnmarshalBlockData(t *testing.T) {
	blockFixture := fixtures.Load("block.json")

	t.Run("unmarshals block data", func(t *testing.T) {
		_, err := UnmarshalBlockData(types.Jsonb{RawMessage: json.RawMessage(blockFixture)})
		if err != nil {
			t.Errorf("should work as expected")
		}
	})

	t.Run("fails to unmarshals block data", func(t *testing.T) {
		_, err := UnmarshalBlockData(types.Jsonb{RawMessage: json.RawMessage(`{"test": 1}`)})
		if err == nil {
			t.Errorf("should work as expected")
		}
	})
}

func Test_SyncableMapper_UnmarshalStateData(t *testing.T) {
	stateFixture := fixtures.Load("state.json")

	t.Run("unmarshals state data", func(t *testing.T) {
		_, err := UnmarshalStateData(types.Jsonb{RawMessage: json.RawMessage(stateFixture)})
		if err != nil {
			t.Errorf("should work as expected")
		}
	})

	t.Run("fails to unmarshals state data", func(t *testing.T) {
		_, err := UnmarshalStateData(types.Jsonb{RawMessage: json.RawMessage(`{"test": 1}`)})
		if err == nil {
			t.Errorf("should work as expected")
		}
	})
}

func Test_SyncableMapper_UnmarshalValidatorsData(t *testing.T) {
	validatorsFixture := fixtures.Load("validators.json")

	t.Run("unmarshals validators data", func(t *testing.T) {
		_, err := UnmarshalValidatorsData(types.Jsonb{RawMessage: json.RawMessage(validatorsFixture)})
		if err != nil {
			t.Errorf("should work as expected")
		}
	})

	t.Run("fails to unmarshals validators data", func(t *testing.T) {
		_, err := UnmarshalValidatorsData(types.Jsonb{RawMessage: json.RawMessage(`{"test": 1}`)})
		if err == nil {
			t.Errorf("should work as expected")
		}
	})
}

func Test_SyncableMapper_UnmarshalTransactionsData(t *testing.T) {
	transactionsFixture := fixtures.Load("transactions.json")

	t.Run("unmarshals transactions data", func(t *testing.T) {
		_, err := UnmarshalTransactionsData(types.Jsonb{RawMessage: json.RawMessage(transactionsFixture)})
		if err != nil {
			t.Errorf("should work as expected")
		}
	})

	t.Run("fails to unmarshals transactions data", func(t *testing.T) {
		_, err := UnmarshalTransactionsData(types.Jsonb{RawMessage: json.RawMessage(`{"test": 1}`)})
		if err == nil {
			t.Errorf("should work as expected")
		}
	})
}


