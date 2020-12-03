package indexer

import (
	"context"
	"reflect"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/validator/validatorpb"

	"github.com/golang/mock/gomock"
)

func TestValidatorSeqCreator_Run(t *testing.T) {
	const syncHeight int64 = 20

	syncTime := *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC))

	seq := &model.Sequence{
		Height: syncHeight,
		Time:   syncTime,
	}

	tests := []struct {
		description string
		raw         []*validatorpb.Validator
		expect      []model.ValidatorSeq
		expectErr   bool
	}{
		{
			description: "updates payload.ValidatorSequences",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "100"},
			},
			expect: []model.ValidatorSeq{
				{
					Sequence:      seq,
					StashAccount:  "validator1",
					ActiveBalance: types.NewQuantityFromInt64(100),
				},
			},

			expectErr: false,
		},
		{
			description: "return error if sequence is invalid",
			raw: []*validatorpb.Validator{
				{StashAccount: "validator1", Balance: "foood"},
			},
			expect:    []model.ValidatorSeq{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorSeq(ctrl)
			task := NewValidatorSeqCreatorTask(dbMock)

			pl := &payload{
				CurrentHeight: syncHeight,
				Syncable: &model.Syncable{
					Height: syncHeight,
					Time:   syncTime,
				},
				RawValidators: tt.raw,
			}

			if err := task.Run(ctx, pl); err != nil && !tt.expectErr {
				t.Errorf("unexpected error, want %v; got %v", tt.expectErr, err)
				return
			}

			// skip payload check if there's an error
			if tt.expectErr {
				return
			}

			if len(pl.ValidatorSequences) != (len(tt.raw)) {
				t.Errorf("expected payload.ValidatorSequences to contain all validator seqs, got: %v; want: %v", len(pl.ValidatorSequences), len(tt.raw))
				return
			}

			for _, expectVal := range tt.expect {
				var found bool
				for _, val := range pl.ValidatorSequences {
					if val.StashAccount == expectVal.StashAccount {
						if !reflect.DeepEqual(val, expectVal) {
							t.Errorf("unexpected entry in payload.ValidatorSequences, got: %v; want: %v", val, expectVal)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing entry in payload.ValidatorSequences, want: %v", expectVal)
				}
			}
		})
	}
}

func TestRewardEraSeqCreatorTask_Run(t *testing.T) {
	const currEra int64 = 20
	tests := []struct {
		description   string
		lastInEra     bool
		validators    ParsedValidatorsData
		expectedKinds []model.RewardKind
	}{
		{description: "updates payload with commission and reward events",
			lastInEra: true,
			validators: ParsedValidatorsData{
				"addr": {
					UnclaimedCommission: "300",
					UnclaimedReward:     "300",
				},
			},
			expectedKinds: []model.RewardKind{model.RewardCommission, model.RewardReward},
		},
		{description: "updates payload with reward events from staker",
			lastInEra: true,
			validators: ParsedValidatorsData{
				"addr": {
					UnclaimedStakerRewards: []stakerReward{{Stash: "AAA", Amount: "123"}, {Stash: "BBB", Amount: "123"}},
				},
			},
			expectedKinds: []model.RewardKind{model.RewardReward, model.RewardReward},
		},
		{description: "does not create rewards if not last in era",
			validators: ParsedValidatorsData{
				"addr": {
					UnclaimedStakerRewards: []stakerReward{{Stash: "AAA", Amount: "123"}, {Stash: "BBB", Amount: "123"}},
				},
			},
			expectedKinds: []model.RewardKind{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			dbMock := mock.NewMockSyncables(ctrl)

			task := NewRewardEraSeqCreatorTask(nil, dbMock)

			pl := &payload{
				ParsedValidators: tt.validators,
				Syncable:         &model.Syncable{Era: currEra, LastInEra: tt.lastInEra},
			}

			if tt.lastInEra {
				dbMock.EXPECT().FindLastInEra(currEra-1).Return(&model.Syncable{Height: 500}, nil).Times(1)
			}

			if err := task.Run(ctx, pl); err != nil {
				t.Errorf("unexpected error on Run, want %v; got %v", nil, err)
				return
			}

			if len(pl.RewardEraSequences) != len(tt.expectedKinds) {
				t.Errorf("unexpected system event count, want %v; got %v", len(tt.expectedKinds), len(pl.RewardEraSequences))
				return
			}

			for i, kind := range tt.expectedKinds {
				if len(pl.RewardEraSequences) > 0 && pl.RewardEraSequences[i].Kind != kind {
					t.Errorf("unexpected system event kind, want %v; got %v", kind, pl.RewardEraSequences[i].Kind)
				}
			}
		})
	}
}
