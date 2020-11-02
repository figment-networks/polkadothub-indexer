package indexer

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/validator/validatorpb"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestValidatorSeqCreator_Run(t *testing.T) {
	var errTestDbFind = errors.New("dberr")
	const syncHeight int64 = 20

	syncTime := *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC))

	seq := &model.Sequence{
		Height: syncHeight,
		Time:   syncTime,
	}

	tests := []struct {
		description   string
		raw           []*validatorpb.Validator
		existing      []model.ValidatorSeq
		dbErr         error
		expectNew     []model.ValidatorSeq
		expectUpdated []model.ValidatorSeq
		expectErr     error
	}{
		{
			description: "updates payload.NewValidatorSequences",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "100", Commission: "900"},
			},
			existing: []model.ValidatorSeq{},
			dbErr:    nil,
			expectNew: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator1",
					ActiveBalance: types.NewQuantityFromInt64(100),
				},
			},

			expectUpdated: []model.ValidatorSeq{},
			expectErr:     nil,
		},
		{
			description: "updates payload.UpdatedValidatorSequences",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "100", Commission: "900"},
			},
			existing: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator1",
					ActiveBalance: types.NewQuantityFromInt64(200),
				},
			},
			dbErr:     nil,
			expectNew: []model.ValidatorSeq{},
			expectUpdated: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator1",
					ActiveBalance: types.NewQuantityFromInt64(100),
				},
			},

			expectErr: nil,
		},
		{
			description: "updates existing and new validators",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "100", Commission: "900"},
				{StashAccount: "validator2", Balance: "200", Commission: "900"},
				{StashAccount: "validator3", Balance: "300", Commission: "900"},
			},
			existing: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator2",
					ActiveBalance: types.NewQuantityFromInt64(789),
				},
			},
			dbErr: nil,
			expectNew: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator1",
					ActiveBalance: types.NewQuantityFromInt64(100),
				},
				{
					Sequence:      seq,
					StashAccount:  "validator3",
					ActiveBalance: types.NewQuantityFromInt64(300),
				},
			},
			expectUpdated: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator2",
					ActiveBalance: types.NewQuantityFromInt64(200),
				},
			},

			expectErr: nil,
		},
		{
			description: "return error if there's an unexpected database error",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "9288", Commission: "900"},
			},
			existing:      []model.ValidatorSeq{},
			dbErr:         errTestDbFind,
			expectNew:     []model.ValidatorSeq{},
			expectUpdated: []model.ValidatorSeq{},
			expectErr:     errTestDbFind,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.description), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorSeq(ctrl)

			if tt.dbErr != nil {
				dbMock.EXPECT().FindAllByHeight(syncHeight).Return(nil, tt.dbErr).Times(1)
			} else {
				dbMock.EXPECT().FindAllByHeight(syncHeight).Return(tt.existing, nil).Times(1)
			}

			task := NewValidatorSeqCreatorTask(dbMock)

			pl := &payload{
				CurrentHeight: syncHeight,
				Syncable: &model.Syncable{
					Height: syncHeight,
					Time:   syncTime,
				},
				RawValidators: tt.raw,
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("unexpected error, want %v; got %v", tt.expectErr, err)
				return
			}

			// skip payload check if there's an error
			if tt.expectErr != nil {
				return
			}

			if len(pl.NewValidatorSequences) != (len(tt.raw) - len(tt.existing)) {
				t.Errorf("expected payload.NewValidatorSequences to contain all new validator seqs, got: %v; want: %v", len(pl.NewValidatorSequences), (len(tt.raw) - len(tt.existing)))
				return
			}

			for _, expectVal := range tt.expectNew {
				var found bool
				for _, val := range pl.NewValidatorSequences {
					if val.StashAccount == expectVal.StashAccount {
						if !reflect.DeepEqual(val, expectVal) {
							t.Errorf("unexpected entry in payload.NewValidatorSequences, got: %v; want: %v", val, expectVal)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing entry in payload.NewValidatorSequences, want: %v", expectVal)
				}
			}

			if len(pl.UpdatedValidatorSequences) != len(tt.existing) {
				t.Errorf("expected payload.UpdatedValidatorSequences to contain all existing validator seqs, got: %v; want: %v", len(pl.UpdatedValidatorSequences), len(tt.existing))
				return
			}

			for _, expectVal := range tt.expectUpdated {
				var found bool
				for _, val := range pl.UpdatedValidatorSequences {
					if val.StashAccount == expectVal.StashAccount {
						if !reflect.DeepEqual(val, expectVal) {
							t.Errorf("unexpected entry in payload.UpdatedValidatorSequences, got: %v; want: %v", val, expectVal)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing entry in payload.UpdatedValidatorSequences, want: %v", expectVal)
				}
			}
		})
	}
}
