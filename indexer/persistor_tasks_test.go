package indexer

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/golang/mock/gomock"
)

func TestSyncerPersistor_Run(t *testing.T) {
	sync := &model.Syncable{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Now()),
	}
	t.Parallel()

	tests := []struct {
		description string
		expectErr   error
	}{
		{"calls db with syncable", nil},
		{"returns error if database errors", fmt.Errorf("test err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockSyncables(ctrl)

			task := NewSyncerPersistorTask(dbMock)

			pl := &payload{
				CurrentHeight: 20,
				Syncable:      sync,
			}

			dbMock.EXPECT().CreateOrUpdate(sync).Return(tt.expectErr).Times(1)

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestBlockSeqPersistor_Run(t *testing.T) {
	seq := &model.BlockSeq{
		Sequence: &model.Sequence{
			Height: 20,
			Time:   *types.NewTimeFromTime(time.Date(1987, 12, 11, 14, 0, 0, 0, time.UTC)),
		},
		ExtrinsicsCount: 10,
	}

	tests := []struct {
		description string
		expectErr   error
	}{
		{"calls db with block sequence", nil},
		{"returns error if database errors", fmt.Errorf("test err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("[new] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockBlockSeq(ctrl)

			task := NewBlockSeqPersistorTask(dbMock)

			pl := &payload{
				CurrentHeight:    20,
				NewBlockSequence: seq,
			}

			dbMock.EXPECT().CreateSeq(seq).Return(tt.expectErr).Times(1)

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})

		t.Run(fmt.Sprintf("[updated] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockBlockSeq(ctrl)

			task := NewBlockSeqPersistorTask(dbMock)

			pl := &payload{
				CurrentHeight:        20,
				UpdatedBlockSequence: seq,
			}

			dbMock.EXPECT().SaveSeq(seq).Return(tt.expectErr).Times(1)

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestValidatorSessionSeqPersistor_Run(t *testing.T) {
	seqs := []model.ValidatorSessionSeq{
		{SessionSequence: &model.SessionSequence{StartHeight: 20}, StashAccount: "acct1", Online: false},
		{SessionSequence: &model.SessionSequence{StartHeight: 20}, StashAccount: "acct2", Online: true},
		{SessionSequence: &model.SessionSequence{StartHeight: 20}, StashAccount: "acct3", Online: false},
	}

	tests := []struct {
		description   string
		lastInSession bool
		expectErr     error
	}{
		{"doesn't persist if not last in session", false, nil},
		{"calls db with all validator sequences", true, nil},
		{"returns error if database errors", true, fmt.Errorf("db err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("[new] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorSessionSeq(ctrl)

			task := NewValidatorSessionSeqPersistorTask(dbMock)

			pl := &payload{
				Syncable:                     &model.Syncable{LastInSession: tt.lastInSession},
				NewValidatorSessionSequences: seqs,
			}

			if tt.lastInSession {
				for _, s := range seqs {
					createSeq := s
					dbMock.EXPECT().CreateSessionSeq(&createSeq).Return(tt.expectErr).Times(1)
					if tt.expectErr != nil {
						// don't expect any more calls
						break
					}
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})

		t.Run(fmt.Sprintf("[updated] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorSessionSeq(ctrl)

			task := NewValidatorSessionSeqPersistorTask(dbMock)

			pl := &payload{
				Syncable:                         &model.Syncable{LastInSession: tt.lastInSession},
				UpdatedValidatorSessionSequences: seqs,
			}

			if tt.lastInSession {
				for _, s := range seqs {
					saveSeq := s
					dbMock.EXPECT().SaveSessionSeq(&saveSeq).Return(tt.expectErr).Times(1)
					if tt.expectErr != nil {
						// don't expect any more calls
						break
					}
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestValidatorEraSeqPersistor_Run(t *testing.T) {
	seqs := []model.ValidatorEraSeq{
		{EraSequence: &model.EraSequence{StartHeight: 20}, StashAccount: "acct1", Commission: 100},
		{EraSequence: &model.EraSequence{StartHeight: 20}, StashAccount: "acct2", Commission: 200},
		{EraSequence: &model.EraSequence{StartHeight: 20}, StashAccount: "acct3", Commission: 50},
	}

	tests := []struct {
		description string
		lastInEra   bool
		expectErr   error
	}{
		{"doesn't persist if not last in era", false, nil},
		{"calls db with all validator sequences", true, nil},
		{"returns error if database errors", true, fmt.Errorf("db err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("[new] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorEraSeq(ctrl)

			task := NewValidatorEraSeqPersistorTask(dbMock)

			pl := &payload{
				Syncable:                 &model.Syncable{LastInEra: tt.lastInEra},
				NewValidatorEraSequences: seqs,
			}

			if tt.lastInEra {
				for _, s := range seqs {
					createSeq := s
					dbMock.EXPECT().CreateEraSeq(&createSeq).Return(tt.expectErr).Times(1)
					if tt.expectErr != nil {
						// don't expect any more calls
						break
					}
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})

		t.Run(fmt.Sprintf("[updated] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorEraSeq(ctrl)

			task := NewValidatorEraSeqPersistorTask(dbMock)

			pl := &payload{
				Syncable:                     &model.Syncable{LastInEra: tt.lastInEra},
				UpdatedValidatorEraSequences: seqs,
			}

			if tt.lastInEra {
				for _, s := range seqs {
					saveSeq := s
					dbMock.EXPECT().SaveEraSeq(&saveSeq).Return(tt.expectErr).Times(1)
					if tt.expectErr != nil {
						// don't expect any more calls
						break
					}
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestValidatorAggPersistor_Run(t *testing.T) {
	aggs := []model.ValidatorAgg{
		{Aggregate: &model.Aggregate{StartedAtHeight: 10}, StashAccount: "acct1", AccumulatedUptimeCount: 100},
		{Aggregate: &model.Aggregate{StartedAtHeight: 20}, StashAccount: "acct2", AccumulatedUptimeCount: 200},
		{Aggregate: &model.Aggregate{StartedAtHeight: 50}, StashAccount: "acct3", AccumulatedUptimeCount: 50},
	}

	tests := []struct {
		description string
		expectErr   error
	}{
		{"calls db with all validator aggs", nil},
		{"returns error if database errors", fmt.Errorf("db err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("[new] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorAgg(ctrl)

			task := NewValidatorAggPersistorTask(dbMock)

			pl := &payload{
				NewValidatorAggregates: aggs,
			}

			for _, s := range aggs {
				createAgg := s
				dbMock.EXPECT().CreateAgg(&createAgg).Return(tt.expectErr).Times(1)
				if tt.expectErr != nil {
					// don't expect any more calls
					break
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})

		t.Run(fmt.Sprintf("[updated] %v", tt.description), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorAgg(ctrl)

			task := NewValidatorAggPersistorTask(dbMock)

			pl := &payload{
				UpdatedValidatorAggregates: aggs,
			}

			for _, s := range aggs {
				saveAgg := s
				dbMock.EXPECT().SaveAgg(&saveAgg).Return(tt.expectErr).Times(1)
				if tt.expectErr != nil {
					// don't expect any more calls
					break
				}
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestEventSequencePersistor_Run(t *testing.T) {
	seqs := []model.EventSeq{
		{Sequence: &model.Sequence{Height: 10}, Method: "method1"},
		{Sequence: &model.Sequence{Height: 20}, Method: "method2"},
		{Sequence: &model.Sequence{Height: 50}, Method: "method3"},
	}

	tests := []struct {
		description string
		expectErr   error
	}{
		{"calls db with all event seqs", nil},
		{"returns error if database errors", fmt.Errorf("db err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockEventSeq(ctrl)

			task := NewEventSeqPersistorTask(dbMock)

			pl := &payload{
				EventSequences: seqs,
			}

			dbMock.EXPECT().BulkUpsert(seqs).Return(tt.expectErr).Times(1)

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestAccountEraSequencePersistor_Run(t *testing.T) {
	seqs := []model.AccountEraSeq{
		{EraSequence: &model.EraSequence{StartHeight: 10}, StashAccount: "acount1"},
		{EraSequence: &model.EraSequence{StartHeight: 20}, StashAccount: "acount2"},
		{EraSequence: &model.EraSequence{StartHeight: 50}, StashAccount: "acount3"},
	}

	tests := []struct {
		description string
		lastInEra   bool
		expectErr   error
	}{
		{"doesn't persist if not last in era", false, nil},
		{"calls db with all account era seqs", true, nil},
		{"returns error if database errors", true, fmt.Errorf("db err")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockAccountEraSeq(ctrl)

			task := NewAccountEraSeqPersistorTask(dbMock)

			pl := &payload{
				Syncable:            &model.Syncable{LastInEra: tt.lastInEra},
				AccountEraSequences: seqs,
			}

			if tt.lastInEra {
				dbMock.EXPECT().BulkUpsert(pl.AccountEraSequences).Return(tt.expectErr).Times(1)
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}

func TestValidatorSeqPersistor_Run(t *testing.T) {
	seqTime := *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC))

	seqs := []model.ValidatorSeq{
		{Sequence: &model.Sequence{Time: seqTime, Height: 20}, StashAccount: "acct1", ActiveBalance: types.NewQuantityFromInt64(100)},
		{Sequence: &model.Sequence{Time: seqTime, Height: 20}, StashAccount: "acct2", ActiveBalance: types.NewQuantityFromInt64(200)},
		{Sequence: &model.Sequence{Time: seqTime, Height: 20}, StashAccount: "acct3", ActiveBalance: types.NewQuantityFromInt64(300)},
	}

	tests := []struct {
		description string
		expectErr   error
	}{
		{"calls db with all validator sequences", nil},
		{"returns error if database errors", fmt.Errorf("db err")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorSeq(ctrl)

			task := NewValidatorSeqPersistorTask(dbMock)

			pl := &payload{
				ValidatorSequences: seqs,
			}

			dbMock.EXPECT().BulkUpsertSeqs(pl.ValidatorSequences).Return(tt.expectErr).Times(1)

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}
		})
	}
}
