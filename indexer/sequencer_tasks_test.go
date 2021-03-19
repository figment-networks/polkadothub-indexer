package indexer

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
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
	const testValidator = "testValidator"

	tests := []struct {
		description   string
		lastInEra     bool
		validator     parsedValidator
		expectedKinds []model.RewardKind
	}{
		{description: "updates payload with commission and reward events",
			lastInEra: true,
			validator: parsedValidator{parsedRewards: parsedRewards{
				Commission: "300",
				Reward:     "300",
				Era:        currEra,
			}},
			expectedKinds: []model.RewardKind{model.RewardCommission, model.RewardReward},
		},
		{description: "updates payload with commission_and_reward event",
			lastInEra: true,
			validator: parsedValidator{parsedRewards: parsedRewards{
				RewardAndCommission: "300",
				Era:                 currEra,
			}},
			expectedKinds: []model.RewardKind{model.RewardCommissionAndReward},
		},
		{description: "updates payload with reward events from staker",
			lastInEra: true,
			validator: parsedValidator{parsedRewards: parsedRewards{
				StakerRewards: []stakerReward{{Stash: "AAA", Amount: "123"}, {Stash: "BBB", Amount: "123"}},
				Era:           currEra,
			}},
			expectedKinds: []model.RewardKind{model.RewardReward, model.RewardReward},
		},
		{description: "creates rewards if not last in era",
			validator: parsedValidator{parsedRewards: parsedRewards{
				StakerRewards: []stakerReward{{Stash: "AAA", Amount: "123"}, {Stash: "BBB", Amount: "123"}},
				Era:           currEra,
			}},
			expectedKinds: []model.RewardKind{model.RewardReward, model.RewardReward},
		},
		{description: "creates rewards if validator era is different from current era",
			validator: parsedValidator{parsedRewards: parsedRewards{
				StakerRewards: []stakerReward{{Stash: "AAA", Amount: "123"}, {Stash: "BBB", Amount: "123"}},
				Era:           currEra - 1,
			},
			},
			expectedKinds: []model.RewardKind{model.RewardReward, model.RewardReward},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			dbMock := mock.NewMockSyncables(ctrl)

			task := NewRewardEraSeqCreatorTask(nil, nil, dbMock, nil)

			pl := &payload{
				ParsedValidators: ParsedValidatorsData{testValidator: tt.validator},
				Syncable:         &model.Syncable{Era: currEra, LastInEra: tt.lastInEra},
			}

			dbMock.EXPECT().FindLastInEra(currEra-1).Return(&model.Syncable{Height: 500}, nil).Times(1)
			if tt.validator.parsedRewards.Era != currEra {
				dbMock.EXPECT().FindLastInEra(tt.validator.parsedRewards.Era).Return(&model.Syncable{Height: 500}, nil).Times(1)
				dbMock.EXPECT().FindLastInEra(tt.validator.parsedRewards.Era-1).Return(&model.Syncable{Height: 500}, nil).Times(1)
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

	markClaimedTest := []struct {
		description   string
		txs           []*transactionpb.Transaction
		events        []*eventpb.Event
		expectErr     bool
		expectClaimed []RewardsClaim
	}{
		{
			description: "expect claims if there's a payout stakers transaction",
			txs:         []*transactionpb.Transaction{testPayoutStakersTx("v1", 182, "abc", 0)},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectClaimed: []RewardsClaim{{182, "v1", "abc"}},
		},
		{
			description: "expect claims if there's multiple payout stakers transaction",
			txs:         []*transactionpb.Transaction{testPayoutStakersTx("v1", 182, "abc", 0), testPayoutStakersTx("v1", 180, "d", 1), testPayoutStakersTx("v2", 180, "e", 2)},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(1, "v1", "2000"),
				testpbRewardEvent(2, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{182, "v1", "abc"}, {180, "v1", "d"}, {180, "v2", "e"}},
		},
		{
			description: "expect claims only for payout stakers tx",
			txs:         []*transactionpb.Transaction{{Section: "staking", Method: "Foo", IsSuccess: true, Hash: "foo"}, testPayoutStakersTx("v1", 180, "abc", 1)},
			events: []*eventpb.Event{
				testpbRewardEvent(1, "v1", "2000"),
			},
			expectClaimed: []RewardsClaim{{180, "v1", "abc"}},
		},
		{
			description: "expect claims if there's a utility batch transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{
					Method: "batch", Section: "utility",
					ExtrinsicIndex: 0,
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					IsSuccess: true,
					Hash:      "abc",
				},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1", "abc"}, {202, "v1", "abc"}, {202, "v2", "abc"}},
		},
		{
			description: "expect claims if there's a utility batchAll transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{
					Method: "batchAll", Section: "utility",
					ExtrinsicIndex: 0,
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					IsSuccess: true,
					Hash:      "abc",
				},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1", "abc"}, {202, "v1", "abc"}, {202, "v2", "abc"}},
		},
		{
			description: "expect claims if there's a proxy proxy transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{
					Method: "proxy", Section: "proxy",
					ExtrinsicIndex: 0,
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					Hash:      "abc",
					IsSuccess: true},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1", "abc"}, {202, "v1", "abc"}, {202, "v2", "abc"}},
		},
		{
			description: "does not expect claims if tx is not a success",
			txs: []*transactionpb.Transaction{
				{
					Method: "batchAll", Section: "utility",
					ExtrinsicIndex: 0,
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
					},
					Hash:      "abc",
					IsSuccess: false},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectClaimed: []RewardsClaim{},
		},
		{
			description: "expect error if there's a batch transaction with call args in unknown format",
			txs: []*transactionpb.Transaction{
				{
					Method: "batchAll", Section: "utility",
					ExtrinsicIndex: 0,
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `[validatorId: "v1", era: "199"]`},
					},
					Hash:      "abc",
					IsSuccess: true},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectErr: true,
		},
	}

	for _, tt := range markClaimedTest {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			rewardsMock := mock.NewMockRewards(ctrl)
			rewardsMock.EXPECT().GetCount(gomock.Any(), gomock.Any()).Return(int64(1), nil).AnyTimes()

			syncablesMock := mock.NewMockSyncables(ctrl)
			syncablesMock.EXPECT().FindLastInEra(gomock.Any()).Return(&model.Syncable{}, nil).AnyTimes()

			validatorMock := mock.NewMockValidatorEraSeq(ctrl)
			validatorMock.EXPECT().FindByEraAndStashAccount(gomock.Any(), gomock.Any()).Return(&model.ValidatorEraSeq{}, nil).AnyTimes()

			task := NewRewardEraSeqCreatorTask(nil, rewardsMock, syncablesMock, validatorMock)

			pl := &payload{
				Syncable: &model.Syncable{Era: currEra},
				RawBlock: &blockpb.Block{
					Extrinsics: tt.txs,
				},
				RawEvents: tt.events,
			}

			err := task.Run(ctx, pl)
			if err != nil && !tt.expectErr {
				t.Errorf("Unexpected error: got %v", err)
				return
			}

			if len(pl.RewardsClaimed) != len(tt.expectClaimed) {
				t.Errorf("unexpected RewardsClaimed count, want %v; got %v", len(tt.expectClaimed), len(pl.RewardsClaimed))
				return
			}

			if len(pl.RewardEraSequences) != 0 {
				t.Errorf("Unexpected Rewards in payload.RewardEraSequences, got %v", pl.RewardsClaimed)
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

func Test_extractRewards(t *testing.T) {
	tests := []struct {
		description    string
		txIdx          int64
		rawClaimsForTx []RewardsClaim
		rewardArgs     []rewardEventArgs
		expectRewards  []model.RewardEraSeq
		expectErr      bool
	}{
		{
			description:    "expect no rewards if there's no claims",
			rawClaimsForTx: []RewardsClaim{},
		},
		{
			description:    "expect error if there's a claim and no rewardArgs",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			expectErr:      true,
		},
		{
			description:    "expect validator and nominator rewards if there are rewardArgs",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			rewardArgs:     []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1500",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom2",
					Kind:                  model.RewardReward,
					Amount:                "2000",
					Claimed:               true,
					TxHash:                "abc",
				},
			},
		},
		{
			description:    "expect rewards to be created for multiple claims",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}, {102, "v1", "abc"}},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"nom1", "1500"},
				{"v1", "2000"},
				{"nom1", "2500"},
				{"v1", "3000"},
				{"nom1", "3500"},
			},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1500",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "2000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "2500",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "3000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "3500",
					Claimed:               true,
					TxHash:                "abc",
				},
			},
		},
		{
			description:    "expect validator rewards to be created for multiple claims",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}, {102, "v2", "abc"}, {102, "v1", "abc"}},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"v1", "2000"},
				{"v2", "3000"},
				{"v1", "4000"},
			},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "2000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v2",
					StashAccount:          "v2",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "3000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "4000",
					Claimed:               true,
					TxHash:                "abc",
				},
			},
		},
		{
			description:    "expect error if event for validator is missing from rewardArgs (all same)",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}, {102, "v1", "abc"}, {103, "v1", "abc"}},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"v1", "2000"},
				{"v1", "4000"},
			},
			expectErr: true,
		},
		{
			description:    "expect error if event for validator is missing from rewardArgs (one different)",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}, {102, "v2", "abc"}, {103, "v1", "abc"}},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"v1", "2000"},
				{"v1", "4000"},
			},
			expectErr: true,
		},
		{
			description:    "expect nominator rewards if  event for validator is missing from raw rewardArgs but there's only one claim",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			rewardArgs:     []rewardEventArgs{{"nom1", "1000"}, {"nom2", "2000"}},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
					TxHash:                "abc",
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom2",
					Kind:                  model.RewardReward,
					Amount:                "2000",
					Claimed:               true,
					TxHash:                "abc",
				},
			},
		},
		{
			description:    "expect error if raw rewardArgs are in reverse order",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v2", "d"}, {102, "v3", "e"}},
			rewardArgs: []rewardEventArgs{
				{"v3", "1000"},
				{"v2", "2000"},
				{"v1", "4000"},
			},
			expectErr: true,
		},
		{
			description:    "expect error rewardArgs are in scrambled order",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v2", "abc"}, {102, "v3", "abc"}, {102, "v4", "abc"}, {102, "v5", "abc"}},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"v3", "2000"},
				{"v2", "4000"},
				{"v4", "2000"},
				{"v5", "2000"},
			},
			expectErr: true,
		},
	}

	var zero int64
	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			rewardsMock := mock.NewMockRewards(ctrl)
			validatorMock := mock.NewMockValidatorEraSeq(ctrl)
			syncablesMock := mock.NewMockSyncables(ctrl)

			for _, c := range tt.rawClaimsForTx {
				rewardsMock.EXPECT().GetCount(c.ValidatorStash, c.Era).Return(zero, nil).Times(1)
				syncablesMock.EXPECT().FindLastInEra(c.Era).Return(&model.Syncable{Era: c.Era, Height: 9}, nil)
				syncablesMock.EXPECT().FindLastInEra(c.Era-1).Return(&model.Syncable{Era: c.Era - 1, Height: 8}, nil)
			}

			task := NewRewardEraSeqCreatorTask(nil, rewardsMock, syncablesMock, validatorMock)

			rewards, claims, err := task.extractRewards(tt.rawClaimsForTx, tt.rewardArgs)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected run error; got nil")
				}
				return
			} else if err != nil {
				t.Errorf("Unexpected run error; got %v", err)
				return
			}

			if len(claims) != 0 {
				t.Errorf("Expect no claims, got %v", claims)
				return
			}

			if len(rewards) != len(tt.expectRewards) {
				t.Errorf("Unexpected parsedReward.StakerRewards length, want: %v, got: %+v", len(tt.expectRewards), len(rewards))
			}

			for _, want := range tt.expectRewards {
				var found bool
				for _, got := range rewards {
					if got.StashAccount == want.StashAccount && got.ValidatorStashAccount == want.ValidatorStashAccount && got.Era == want.Era {
						found = true
						if got.Amount != want.Amount {
							t.Errorf("Unexpected Amount for %v; want: %v, got: %+v", got.StashAccount, want.Amount, got.Amount)
						}
						if !got.Claimed {
							t.Errorf("Unexpected parsedReward.IsClaimed, want: %v, got: %+v", true, got.Claimed)
						}
						if got.Kind != want.Kind {
							t.Errorf("Unexpected reward kind for %v, want: %v, got: %+v", got.StashAccount, want.Kind, got.Kind)
						}
						if got.TxHash != want.TxHash {
							t.Errorf("Unexpected tx hash for %v, want: %v, got: %+v", got.StashAccount, want.TxHash, got.TxHash)
						}
					}
				}
				if !found {
					t.Errorf("Expected to find entry for %v in rewards; got entries: %v", want, rewards)
				}
			}
		})
	}

	testsForExistingRewards := []struct {
		description         string
		rawClaimsForTx      []RewardsClaim
		returnCountForClaim map[RewardsClaim]int64
		returnValidatorSeq  *model.ValidatorEraSeq
		rewardArgs          []rewardEventArgs
		expectRewards       []model.RewardEraSeq
		expectClaimed       []RewardsClaim
		expectErr           bool
	}{
		{
			description:         "expect claim only for rewards that exist in db",
			rawClaimsForTx:      []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}, {102, "v1", "abc"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1", "abc"}: 0, {101, "v1", "abc"}: 2, {102, "v1", "abc"}: 0},
			rewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"nom1", "1500"},
				{"v1", "2000"},
				{"nom1", "2500"},
				{"v1", "3000"},
				{"nom1", "3500"},
			},
			expectClaimed: []RewardsClaim{{101, "v1", "abc"}},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1500",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "3000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "3500",
					Claimed:               true,
				},
			},
		},
		{
			description:         "expect error if rewards count < len(rewards)",
			rawClaimsForTx:      []RewardsClaim{{100, "v1", "abc"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1", "abc"}: 1},
			rewardArgs:          []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectErr:           true,
		},
		{
			description:         "expect error if rewards count > len(rewards)+2",
			rawClaimsForTx:      []RewardsClaim{{100, "v1", "abc"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1", "abc"}: 5},
			rewardArgs:          []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectErr:           true,
		},
		{
			description:         "expect claim if rewards count == len(rewards)",
			rawClaimsForTx:      []RewardsClaim{{100, "v1", "abc"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1", "abc"}: 3},
			rewardArgs:          []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectClaimed:       []RewardsClaim{{100, "v1", "abc"}},
		},
		{
			description:         "expect claim if rewards count == len(rewards)+1",
			rawClaimsForTx:      []RewardsClaim{{100, "v1", "abc"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1", "abc"}: 4},
			rewardArgs:          []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectClaimed:       []RewardsClaim{{100, "v1", "abc"}},
		},
	}

	for _, tt := range testsForExistingRewards {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			rewardsMock := mock.NewMockRewards(ctrl)
			validatorMock := mock.NewMockValidatorEraSeq(ctrl)
			syncablesMock := mock.NewMockSyncables(ctrl)

			for _, c := range tt.rawClaimsForTx {
				count, ok := tt.returnCountForClaim[c]
				if !ok {
					t.Errorf("Missing entry in tt.returnCountForClaim; want: %v", c)
				}
				rewardsMock.EXPECT().GetCount(c.ValidatorStash, c.Era).Return(count, nil).Times(1)
				syncablesMock.EXPECT().FindLastInEra(c.Era).Return(&model.Syncable{Era: c.Era, Height: 9}, nil)
				syncablesMock.EXPECT().FindLastInEra(c.Era-1).Return(&model.Syncable{Era: c.Era - 1, Height: 8}, nil)
				validatorMock.EXPECT().FindByEraAndStashAccount(c.Era, c.ValidatorStash).Return(tt.returnValidatorSeq, nil).AnyTimes()
			}

			task := NewRewardEraSeqCreatorTask(nil, rewardsMock, syncablesMock, validatorMock)

			rewards, claims, err := task.extractRewards(tt.rawClaimsForTx, tt.rewardArgs)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected run error; got nil")
				}
				return
			} else if err != nil {
				t.Errorf("Unexpected run error; got %v", err)
				return
			}

			if len(rewards) != len(tt.expectRewards) {
				t.Errorf("Unexpected rewards length, want: %v, got: %+v", len(tt.expectRewards), len(rewards))
			}

			if len(claims) != len(tt.expectClaimed) {
				t.Errorf("Unexpected claims length, want: %v, got: %+v", len(tt.expectClaimed), len(claims))
			}

			for _, want := range tt.expectClaimed {
				var found bool
				for _, got := range claims {
					if got.Era == want.Era && got.ValidatorStash == want.ValidatorStash {
						found = true
					}
				}
				if !found {
					t.Errorf("Missing entry %v in claims; got entries: %v", want, claims)
				}
			}
		})
	}

	dbErr := errors.New("dbErr")
	testDbErrs := []struct {
		description          string
		rawClaimsForTx       []RewardsClaim
		rewardArgs           []rewardEventArgs
		returnRewardsDbErr   error
		returnSyncablesDbErr error
		expectErr            error
	}{
		{
			description:    "expect no error when db returns no error",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			rewardArgs:     []rewardEventArgs{{"v1", "1500"}},
		},
		{
			description:        "expect error when rewards db returns error",
			rawClaimsForTx:     []RewardsClaim{{100, "v1", "abc"}},
			rewardArgs:         []rewardEventArgs{{"v1", "1500"}},
			returnRewardsDbErr: dbErr,
			expectErr:          dbErr,
		},
		{
			description:          "expect error when syncables db returns error",
			rawClaimsForTx:       []RewardsClaim{{100, "v1", "abc"}},
			rewardArgs:           []rewardEventArgs{{"v1", "1500"}},
			returnSyncablesDbErr: dbErr,
			expectErr:            dbErr,
		},
	}

	for _, tt := range testDbErrs {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			rewardsMock := mock.NewMockRewards(ctrl)
			validatorMock := mock.NewMockValidatorEraSeq(ctrl)
			syncablesMock := mock.NewMockSyncables(ctrl)

			rewardsMock.EXPECT().GetCount(gomock.Any(), gomock.Any()).Return(zero, tt.returnRewardsDbErr).AnyTimes()
			syncablesMock.EXPECT().FindLastInEra(gomock.Any()).Return(&model.Syncable{}, tt.returnSyncablesDbErr).AnyTimes()

			task := NewRewardEraSeqCreatorTask(nil, rewardsMock, syncablesMock, validatorMock)

			_, _, err := task.extractRewards(tt.rawClaimsForTx, tt.rewardArgs)
			if err != tt.expectErr {
				t.Errorf("Unexpected error on run; got %v, want: %v", err, tt.expectErr)
				return
			}

		})
	}
}

func Test_getLegitimateClaimsAndRewardArgs(t *testing.T) {
	tests := []struct {
		description      string
		txIdx            int64
		rawClaimsForTx   []RewardsClaim
		events           []*eventpb.Event
		expectRewardArgs []rewardEventArgs
		expectClaims     []RewardsClaim
		expectErr        bool
	}{
		{
			description:    "expect no results if there's no claims",
			rawClaimsForTx: []RewardsClaim{},
			events:         []*eventpb.Event{},
		},
		{
			description:    "expect no claims or reward args if there's no reward events",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			events:         []*eventpb.Event{{Section: sectionStaking, Method: "Foo"}},
		},
		{
			description:      "expect validator and nominator rewardargs if there are reward events",
			rawClaimsForTx:   []RewardsClaim{{100, "v1", "abc"}},
			events:           []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectRewardArgs: []rewardEventArgs{{"v1", "1500"}, {"nom1", "1000"}, {"nom2", "2000"}},
			expectClaims:     []RewardsClaim{{100, "v1", "abc"}},
		},
		{
			description:    "expect no rewardargs  from non-reward events",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			events: []*eventpb.Event{
				{Section: sectionStaking, Method: "Foo"},
				testpbRewardEvent(0, "v1", "1500"),
				testpbRewardEvent(0, "nom1", "1000"),
				{Section: "Foo", Method: "Foo"},
			},
			expectClaims: []RewardsClaim{{100, "v1", "abc"}},
			expectRewardArgs: []rewardEventArgs{
				{"v1", "1500"},
				{"nom1", "1000"},
			},
		},
		{
			description:    "expect rewardargs only from events from same transaction",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}},
			txIdx:          1,
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v0", "2000"),
				testpbRewardEvent(0, "nom1", "2000"),
				testpbRewardEvent(1, "v1", "1500"),
				testpbRewardEvent(1, "nom1", "1000"),
				testpbRewardEvent(2, "v1", "5000"),
				testpbRewardEvent(2, "nom1", "5000"),
				{ExtrinsicIndex: 2, Section: "Foo", Method: "Foo"},
			},
			expectClaims: []RewardsClaim{{100, "v1", "abc"}},
			expectRewardArgs: []rewardEventArgs{
				{"v1", "1500"},
				{"nom1", "1000"},
			},
		},
		{
			description:    "expect reward args to be created for multiple claims",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v2", "abc"}, {102, "v3", "abc"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "nom1", "1500"),
				testpbRewardEvent(0, "v2", "2000"),
				testpbRewardEvent(0, "nom1", "2500"),
				testpbRewardEvent(0, "v3", "3000"),
				testpbRewardEvent(0, "nom1", "3500"),
			},
			expectClaims: []RewardsClaim{{100, "v1", "abc"}, {101, "v2", "abc"}, {102, "v3", "abc"}},
			expectRewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"nom1", "1500"},
				{"v2", "2000"},
				{"nom1", "2500"},
				{"v3", "3000"},
				{"nom1", "3500"},
			},
		},
		{
			description:    "expect reward args to be created for multiple claims when same validator is in list",
			rawClaimsForTx: []RewardsClaim{{100, "v3", "abc"}, {101, "v1", "abc"}, {102, "v2", "abc"}, {102, "v1", "abc"}},
			events: []*eventpb.Event{
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v3", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "nom1", "1500"),
				testpbRewardEvent(0, "v2", "3000"),
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v1", "4000"),
			},
			expectClaims: []RewardsClaim{{100, "v3", "abc"}, {101, "v1", "abc"}, {102, "v2", "abc"}, {102, "v1", "abc"}},
			expectRewardArgs: []rewardEventArgs{
				{"v3", "1000"},
				{"v1", "2000"},
				{"nom1", "1500"},
				{"v2", "3000"},
				{"v1", "4000"},
			},
		},
		{
			description:    "expect claim to be filtered if no reward events exists for validator",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v2", "abc"}, {102, "v1", "abc"}, {103, "v1", "abc"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v1", "4000"),
			},
			expectClaims: []RewardsClaim{{100, "v1", "abc"}, {102, "v1", "abc"}, {103, "v1", "abc"}},
			expectRewardArgs: []rewardEventArgs{
				{"v1", "1000"},
				{"v1", "2000"},
				{"v1", "4000"},
			},
		},
		{
			description:    "expect error if no reward events exist for a validator but there's a validator claim with events for same validator",
			rawClaimsForTx: []RewardsClaim{{100, "v1", "abc"}, {101, "v1", "abc"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectErr: true,
		},
	}

	// var zero int64
	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			validatorMock := mock.NewMockValidatorEraSeq(ctrl)

			for _, c := range tt.rawClaimsForTx {
				validatorMock.EXPECT().FindByEraAndStashAccount(c.Era, c.ValidatorStash).Return(&model.ValidatorEraSeq{}, nil).Times(1)
			}

			task := NewRewardEraSeqCreatorTask(nil, nil, nil, validatorMock)

			legitimateClaims, rewardArgs, err := task.getLegitimateClaimsAndRewardArgs(tt.rawClaimsForTx, tt.events, tt.txIdx)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected run error; got nil")
				}
				return
			} else if err != nil {
				t.Errorf("Unexpected run error; got %v", err)
				return
			}

			if len(legitimateClaims) != len(tt.expectClaims) {
				t.Errorf("Unexpected claims length, want: %v, got: %+v", len(tt.expectClaims), len(legitimateClaims))
				return
			}
			for i, want := range tt.expectClaims {
				got := legitimateClaims[i]
				if got.ValidatorStash != want.ValidatorStash || got.Era != want.Era {
					t.Errorf("Unexpected claim at index %v; want: %v, got: %+v", i, want, got)
				}
			}

			if len(rewardArgs) != len(tt.expectRewardArgs) {
				t.Errorf("Unexpected rewardsArgs length, want: %v, got: %+v", len(tt.expectRewardArgs), len(rewardArgs))
				return
			}
			for i, want := range tt.expectRewardArgs {
				got := rewardArgs[i]
				if got.stash != want.stash || got.amount != want.amount {
					t.Errorf("Unexpected rewardarg at index %v; want: %v, got: %+v", i, want, got)
				}
			}
		})
	}
}

func testpbRewardEvent(txIdx int64, stash, amount string) *eventpb.Event {
	return &eventpb.Event{ExtrinsicIndex: txIdx, Method: "Reward", Section: "staking", Data: []*eventpb.EventData{{Name: "AccountId", Value: stash}, {Name: "Balance", Value: amount}}}
}

func testPayoutStakersTx(stash string, era int64, hash string, idx int64) *transactionpb.Transaction {
	return &transactionpb.Transaction{
		ExtrinsicIndex: idx,
		Hash:           hash,
		Method:         txMethodPayoutStakers,
		Section:        sectionStaking,
		IsSuccess:      true,
		Args:           fmt.Sprintf(`["%v","%v"]`, stash, era),
	}
}
