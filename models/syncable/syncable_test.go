package syncable

import (
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
	"testing"
	"time"
)

func Test_Report(t *testing.T) {
	chainId := "chain123"
	model := &shared.Model{}
	seq := &shared.Sequence{
		ChainId: chainId,
		Height:  types.Height(10),
		Time:    *types.NewTimeFromTime(time.Now()),
	}

	t.Run("validation success", func(t *testing.T) {
		m := Model{
			Model: model,
			Sequence: seq,

			Type: types.SyncableTypeBlock,
		}

		if !m.Valid() {
			t.Errorf("model should be valid %+v", m)
		}
	})

	t.Run("validation failed", func(t *testing.T) {
		m := Model{
			Model: model,
			Sequence: seq,
		}

		if m.Valid() {
			t.Errorf("model should not be valid %+v", m)
		}
	})

	t.Run("MarkProcessed()", func(t *testing.T) {
		m := Model{
			Model: model,
			Sequence: seq,

			Type: types.SyncableTypeBlock,
		}

		m.MarkProcessed(types.ID(10))

		if *m.ReportID != types.ID(10) ||
			m.ProcessedAt.IsZero() {
			t.Errorf( "values not updated correctly %+v", m)
		}
	})
}


