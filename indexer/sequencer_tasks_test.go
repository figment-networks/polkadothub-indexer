package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/validator/validatorpb"
	"github.com/pkg/errors"

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

func TestClaimedRewardEraSeqCreatorTask_Run(t *testing.T) {
	testValidator := "validator_stash1"

	txtests := []struct {
		description   string
		txs           []model.TransactionSeq
		expectErr     error
		expectClaimed []RewardsClaim
	}{
		{
			description:   "updates payload if there's a payout stakers transaction",
			txs:           []model.TransactionSeq{testPayoutStakersTx(testValidator, 182)},
			expectClaimed: []RewardsClaim{{182, testValidator}},
		},
		{
			description:   "updates payload if there's multiple payout stakers transaction",
			txs:           []model.TransactionSeq{testPayoutStakersTx(testValidator, 182), testPayoutStakersTx(testValidator, 180)},
			expectClaimed: []RewardsClaim{{182, testValidator}, {180, testValidator}},
		},
		{
			description: "does not update payload if there's no payout stakers transaction",
			txs:         []model.TransactionSeq{{Section: "staking", Method: "Foo"}},
		},
	}

	for _, tt := range txtests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockSyncables(ctrl)
			rewardsMock := mock.NewMockRewards(ctrl)

			rewardsMock.EXPECT().GetCount(gomock.Any(), gomock.Any()).Return(int64(1), nil).AnyTimes()

			task := NewClaimedRewardEraSeqCreatorTask(nil, rewardsMock, dbMock, nil)

			pl := &payload{
				TransactionSequences: tt.txs,
			}

			if err := task.Run(ctx, pl); err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}

			if len(pl.RewardsClaimed) != len(tt.expectClaimed) {
				t.Errorf("unexpected RewardsClaimed count, want %v; got %v", len(tt.expectClaimed), len(pl.RewardsClaimed))
				return
			}

			for i, expect := range tt.expectClaimed {
				if pl.RewardsClaimed[i] != expect {
					t.Errorf("unexpected rewards claim, want %v; got %v", expect, pl.RewardsClaimed[i])
				}
			}

		})
	}
}

func Test_getRewardsFromEvents(t *testing.T) {
	testValidator := "validator_stash1"
	var testEra int64 = 182
	var dbErr = errors.New("test err")

	txtests := []struct {
		description                 string
		events                      []model.EventSeq
		validatorEraSeq             *model.ValidatorEraSeq
		validatorDbErr              error
		expectErr                   error
		expectRewardCount           int
		expectCommissionCount       int
		expectRewardCommissionCount int
	}{
		{
			description: "expect no rewards if there's no events",
			events:      []model.EventSeq{},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
		},
		{
			description: "expect reward event from nominators",
			events:      []model.EventSeq{testRewardEvent(t, "nom1", "2000"), testRewardEvent(t, "nom2", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				StakersStake: types.NewQuantityFromInt64(400),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectRewardCount: 2,
		},
		{
			description: "expect reward event from validator",
			events:      []model.EventSeq{testRewardEvent(t, testValidator, "400"), testRewardEvent(t, "nom1", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectRewardCount: 2,
		},
		{
			description: "expect only validator commission when commission = 100%",
			events:      []model.EventSeq{testRewardEvent(t, testValidator, "2000")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   1000000000,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectRewardCount:     0,
			expectCommissionCount: 1,
		},
		{
			description: "expect only validator commission when validator is not staked",
			events:      []model.EventSeq{testRewardEvent(t, testValidator, "2000")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   500000000,
				StakersStake: types.NewQuantityFromInt64(400),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectRewardCount:     0,
			expectCommissionCount: 1,
		},
		{
			description: "expect validator commission_and_reward when validator has commission and is staked",
			events:      []model.EventSeq{testRewardEvent(t, testValidator, "2000"), testRewardEvent(t, "nom1", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   500000000,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectRewardCount:           1,
			expectRewardCommissionCount: 1,
		},
		{
			description: "expect err if db errors",
			events:      []model.EventSeq{testRewardEvent(t, testValidator, "2000"), testRewardEvent(t, "nom1", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   500000000,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			validatorDbErr: dbErr,
			expectErr:      dbErr,
		},
	}

	for _, tt := range txtests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			syncablesMock := mock.NewMockSyncables(ctrl)
			validatorMock := mock.NewMockValidatorEraSeq(ctrl)

			syncablesMock.EXPECT().FindLastInEra(testEra-1).Return(&model.Syncable{}, nil).Times(1)
			syncablesMock.EXPECT().FindLastInEra(testEra).Return(&model.Syncable{}, nil).Times(1)

			validatorMock.EXPECT().FindByEraAndStashAccount(testEra, testValidator).Return(tt.validatorEraSeq, tt.validatorDbErr).Times(1)

			task := NewClaimedRewardEraSeqCreatorTask(nil, nil, syncablesMock, validatorMock)

			rewards, err := task.getRewardsFromEvents(testValidator, testEra, tt.events)
			if err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}

			var rewardCount, commissionCount, rewardCommissionCount int
			for _, reward := range rewards {
				if reward.Kind == model.RewardCommission {
					commissionCount++
				} else if reward.Kind == model.RewardReward {
					rewardCount++
				} else {
					rewardCommissionCount++
				}
			}

			if rewardCount != tt.expectRewardCount {
				t.Errorf("unexpected  reward count, want %v; got %v", tt.expectRewardCount, rewardCount)
				return
			}
			if commissionCount != tt.expectCommissionCount {
				t.Errorf("unexpected commission reward count, want %v; got %v", tt.expectCommissionCount, commissionCount)
				return
			}
			if rewardCommissionCount != tt.expectRewardCommissionCount {
				t.Errorf("unexpected reward-and-commission reward count, want %v; got %v", tt.expectRewardCommissionCount, rewardCommissionCount)
				return
			}
		})
	}
}

func testRewardEvent(t *testing.T, stash, amount string) model.EventSeq {
	eventData, err := json.Marshal([]model.EventData{{Name: "AccountId", Value: stash}, {Name: "Balance", Value: amount}})
	if err != nil {
		t.Fatalf("unexpected err while marshalling json: %v", err)
	}
	return model.EventSeq{Method: "Reward", Section: "staking", Data: types.Jsonb{RawMessage: eventData}}
}

func testPayoutStakersTx(stash string, era int64) model.TransactionSeq {
	return model.TransactionSeq{
		Method:  model.TxMethodPayoutStakers,
		Section: model.TxSectionStaking,
		Args:    fmt.Sprintf("[\"%v\",\"%v\"]", stash, era),
	}
}
