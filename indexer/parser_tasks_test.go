package indexer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	mock_client "github.com/figment-networks/polkadothub-indexer/mock/client"
	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"

	"github.com/golang/mock/gomock"
)

func TestBlockParserTask_Run(t *testing.T) {
	tests := []struct {
		description         string
		rawBlock            *blockpb.Block
		expectedParsedBlock ParsedBlockData
	}{
		{"updates ParsedBlockData with signed extrinsic",
			&blockpb.Block{
				Extrinsics: []*transactionpb.Transaction{
					{IsSignedTransaction: true},
				},
			},
			ParsedBlockData{
				ExtrinsicsCount:         1,
				UnsignedExtrinsicsCount: 0,
				SignedExtrinsicsCount:   1,
			},
		},
		{"updates ParsedBlockData with unsigned extrinsic",
			&blockpb.Block{
				Extrinsics: []*transactionpb.Transaction{
					{IsSignedTransaction: false},
				},
			},
			ParsedBlockData{
				ExtrinsicsCount:         1,
				UnsignedExtrinsicsCount: 1,
				SignedExtrinsicsCount:   0,
			},
		},
		{"updates ParsedBlockData with multiple extrinsics",
			&blockpb.Block{
				Extrinsics: []*transactionpb.Transaction{
					{IsSignedTransaction: false},
					{IsSignedTransaction: false},
					{IsSignedTransaction: false},
					{IsSignedTransaction: true},
					{IsSignedTransaction: false},
				},
			},
			ParsedBlockData{
				ExtrinsicsCount:         5,
				UnsignedExtrinsicsCount: 4,
				SignedExtrinsicsCount:   1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // need to set this since running tests in parallel
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			task := NewBlockParserTask()

			pl := &payload{
				RawBlock: tt.rawBlock,
			}

			if err := task.Run(ctx, pl); err != nil {
				t.Errorf("unexpected error on Run, want %v; got %v", nil, err)
				return
			}

			if pl.ParsedBlock != tt.expectedParsedBlock {
				t.Errorf("Unexpected ParsedBlock, want: %+v, got: %+v", tt.expectedParsedBlock, pl.ParsedBlock)
				return
			}
		})
	}
}

func TestValidatorParserTask_Run(t *testing.T) {
	name1 := "validator1"
	staking1 := stakingpb.Validator{StashAccount: name1, Commission: 100}
	performance1 := validatorperformancepb.Validator{StashAccount: name1, Online: true}

	name2 := "validator2"
	staking2 := stakingpb.Validator{StashAccount: name2, Commission: 200}
	performance2 := validatorperformancepb.Validator{StashAccount: name2, Online: false}

	tests := []struct {
		description              string
		rawStakingState          *stakingpb.Staking
		rawValidatorPerformances []*validatorperformancepb.Validator
		expectedParsedValidators ParsedValidatorsData
	}{
		{"updates empty state",
			&stakingpb.Staking{},
			[]*validatorperformancepb.Validator{},
			ParsedValidatorsData{},
		},
		{"updates ParsedValidator with staking and performance data",
			&stakingpb.Staking{
				Validators: []*stakingpb.Validator{&staking1, &staking2},
			},
			[]*validatorperformancepb.Validator{&performance1, &performance2},
			ParsedValidatorsData{
				name1: {
					Staking:     &staking1,
					Performance: &performance1,
				},
				name2: {
					Staking:     &staking2,
					Performance: &performance2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			mockClient := mock_client.NewMockAccountClient(ctrl)
			for _, validator := range tt.rawStakingState.GetValidators() {
				mockClient.EXPECT().GetIdentity(validator.StashAccount).Return(&accountpb.GetIdentityResponse{Identity: &accountpb.AccountIdentity{DisplayName: ""}}, nil)
			}

			task := NewValidatorsParserTask(nil, mockClient, nil, nil, nil)
			pl := &payload{
				RawStaking:              tt.rawStakingState,
				RawValidatorPerformance: tt.rawValidatorPerformances,
			}

			if err := task.Run(ctx, pl); err != nil {
				t.Errorf("unexpected error on Run, want %v; got %v", nil, err)
				return
			}

			if len(pl.ParsedValidators) != len(tt.expectedParsedValidators) {
				t.Errorf("Unexpected ParsedValidators entry length, want: %+v, got: %+v", len(tt.expectedParsedValidators), len(pl.ParsedValidators))
				return
			}

			for key, expected := range tt.expectedParsedValidators {
				got := pl.ParsedValidators[key]
				if got.Staking != expected.Staking {
					t.Errorf("Unexpected Staking, want: %+v, got: %+v", expected.Staking, got.Staking)
					return
				}
				if got.Performance != expected.Performance {
					t.Errorf("Unexpected Performance, want: %+v, got: %+v", expected.Performance, got.Performance)
					return
				}
			}
		})
	}

	var syncableEra int64 = 100
	parsedUnclaimedRewardTests := []struct {
		description            string
		rawValidator           *stakingpb.Validator
		totalRewardPoints      int64
		totalRewardPayout      string
		expectCommission       bool
		expectReward           bool
		expectEra              int64
		expectNumStakerRewards int
	}{
		{description: "updates ParsedValidators with reward events",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   30000000,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 10, IsRewardEligible: true},
					{StashAccount: "B", Stake: 10, IsRewardEligible: true},
				},
			},
			totalRewardPoints:      100,
			totalRewardPayout:      "4000",
			expectCommission:       true,
			expectEra:              syncableEra,
			expectNumStakerRewards: 2,
		},
		{description: "does not update ParsedRewards if there's no reward payout",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   30000000,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 10, IsRewardEligible: true},
					{StashAccount: "B", Stake: 10, IsRewardEligible: true},
				},
			},
			totalRewardPoints: 100,
			totalRewardPayout: "",
		},
		{description: "Does not create staker rewards if commission is 100%",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   1000000000,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 10, IsRewardEligible: true},
					{StashAccount: "B", Stake: 10, IsRewardEligible: true},
				},
			},
			totalRewardPoints: 100,
			totalRewardPayout: "4000",
			expectEra:         syncableEra,
			expectCommission:  true,
		},
		{description: "Does not create commission if commission is 100%",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   0,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 10, IsRewardEligible: true},
					{StashAccount: "B", Stake: 10, IsRewardEligible: true},
				},
			},
			totalRewardPoints:      100,
			totalRewardPayout:      "4000",
			expectEra:              syncableEra,
			expectNumStakerRewards: 2,
		},
		{description: "Does not create staker reward if reward is 0",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   300000000,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 20, IsRewardEligible: true},
					{StashAccount: "B", Stake: 0, IsRewardEligible: true},
				},
			},
			totalRewardPoints:      100,
			totalRewardPayout:      "4000",
			expectCommission:       true,
			expectEra:              syncableEra,
			expectNumStakerRewards: 1,
		},
		{description: "expect validtor reward if validator is staked",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   300000000,
				StashAccount: name1,
				TotalStake:   20,
				OwnStake:     10,
			},
			totalRewardPoints: 100,
			totalRewardPayout: "4000",
			expectEra:         syncableEra,
			expectCommission:  true,
			expectReward:      true,
		},
		{description: "Does not create staker reward if staker is ineligible",
			rawValidator: &stakingpb.Validator{
				RewardPoints: 50,
				Commission:   300000000,
				StashAccount: name1,
				TotalStake:   20,
				Stakers: []*stakingpb.Stake{
					{StashAccount: "A", Stake: 20, IsRewardEligible: false},
				},
			},
			totalRewardPoints: 100,
			totalRewardPayout: "4000",
			expectEra:         syncableEra,
			expectCommission:  true,
		},
	}

	for _, tt := range parsedUnclaimedRewardTests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			mockClient := mock_client.NewMockAccountClient(ctrl)
			mockClient.EXPECT().GetIdentity(gomock.Any()).Return(nil, nil)

			task := NewValidatorsParserTask(nil, mockClient, nil, nil, nil)

			pl := &payload{
				Syncable: &model.Syncable{Era: syncableEra},
				RawStaking: &stakingpb.Staking{
					TotalRewardPayout: tt.totalRewardPayout,
					TotalRewardPoints: tt.totalRewardPoints,
					Validators:        []*stakingpb.Validator{tt.rawValidator},
				},
			}

			if err := task.Run(ctx, pl); err != nil {
				t.Errorf("unexpected error on Run, want %v; got %v", nil, err)
				return
			}

			if len(pl.ParsedValidators) != 1 {
				t.Errorf("Unexpect ParsedValidators entry length, want: %+v, got: %+v", 1, len(pl.ParsedValidators))
				return
			}

			for _, validator := range pl.ParsedValidators {
				got := validator.parsedRewards

				if tt.expectEra != got.Era {
					t.Errorf("Unexpected Era, want: %v, got: %+v", tt.expectEra, got.Era)
					return
				}

				if got.IsClaimed {
					t.Errorf("Unexpected IsClaimed, want: %v, got: %+v", false, got.IsClaimed)
					return
				}

				if tt.expectCommission != (got.Commission != "") {
					t.Errorf("Unexpected Commission, got: %+v", got.Commission)
					return
				}

				if tt.expectReward != (got.Reward != "") {
					t.Errorf("Unexpected Reward, got: %+v", got.Reward)
					return
				}

				if tt.expectReward != (got.Reward != "") {
					t.Errorf("Unexpected Reward, got: %+v", got.Reward)
					return
				}

				if tt.expectNumStakerRewards != len(got.StakerRewards) {
					t.Errorf("Unexpected StakerRewards, want: %v, got: %+v", tt.expectNumStakerRewards, len(got.StakerRewards))
					return
				}
			}
		})
	}
}

func TestTransactionParserTask_Run(t *testing.T) {
	markClaimedTest := []struct {
		description   string
		txs           []*transactionpb.Transaction
		events        []*eventpb.Event
		expectErr     bool
		expectClaimed []RewardsClaim
	}{
		{
			description: "expect claims if there's a payout stakers transaction",
			txs:         []*transactionpb.Transaction{testPayoutStakersTx("v1", 182)},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectClaimed: []RewardsClaim{{182, "v1"}},
		},
		{
			description: "expect claims if there's multiple payout stakers transaction",
			txs:         []*transactionpb.Transaction{testPayoutStakersTx("v1", 182), testPayoutStakersTx("v1", 180), testPayoutStakersTx("v2", 180)},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(1, "v1", "2000"),
				testpbRewardEvent(2, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{182, "v1"}, {180, "v1"}, {180, "v2"}},
		},
		{
			description: "expect claims only for payout stakers tx",
			txs:         []*transactionpb.Transaction{{Section: "staking", Method: "Foo", IsSuccess: true}, testPayoutStakersTx("v1", 180)},
			events: []*eventpb.Event{
				testpbRewardEvent(1, "v1", "2000"),
			},
			expectClaimed: []RewardsClaim{{180, "v1"}},
		},
		{
			description: "expect claims if there's a utility batch transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{Method: "batch", Section: "utility",
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					IsSuccess: true},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1"}, {202, "v1"}, {202, "v2"}},
		},
		{
			description: "expect claims if there's a utility batchAll transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{Method: "batchAll", Section: "utility",
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					IsSuccess: true},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1"}, {202, "v1"}, {202, "v2"}},
		},
		{
			description: "expect claims if there's a proxy proxy transaction containing payout stakers txs",
			txs: []*transactionpb.Transaction{
				{Method: "proxy", Section: "proxy",
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v1","202"]`},
						{Method: "payoutStakers", Section: "staking", Value: `["v2","202"]`},
					},
					IsSuccess: true},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "2000"),
			},
			expectClaimed: []RewardsClaim{{199, "v1"}, {202, "v1"}, {202, "v2"}},
		},
		{
			description: "does not expect claims if tx is not a success",
			txs: []*transactionpb.Transaction{
				{Method: "batchAll", Section: "utility",
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `["v1","199"]`},
					}, IsSuccess: false},
			},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
			},
			expectClaimed: []RewardsClaim{},
		},
		{
			description: "expect error if there's a batch transaction with call args in unknown format",
			txs: []*transactionpb.Transaction{
				{Method: "batchAll", Section: "utility",
					CallArgs: []*transactionpb.CallArg{
						{Method: "payoutStakers", Section: "staking", Value: `[validatorId: "v1", era: "199"]`},
					},
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

			task := NewTransactionParserTask(nil, nil, rewardsMock, syncablesMock, nil)

			pl := &payload{
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

func Test_addRewardsFromEvents(t *testing.T) {
	tests := []struct {
		description    string
		txIdx          int64
		rawClaimsForTx []RewardsClaim
		events         []*eventpb.Event
		expectRewards  []model.RewardEraSeq
		expectErr      bool
	}{
		{
			description:    "expect no rewards if there's no claims",
			rawClaimsForTx: []RewardsClaim{},
			events:         []*eventpb.Event{},
		},
		{
			description:    "expect error if there's a claim and no events",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events:         []*eventpb.Event{},
			expectErr:      true,
		},
		{
			description:    "expect validator and nominator rewards if there are reward events",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events:         []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1500",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom2",
					Kind:                  model.RewardReward,
					Amount:                "2000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect no rewards  from non-reward events",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events: []*eventpb.Event{
				{Section: sectionStaking, Method: "Foo"},
				testpbRewardEvent(0, "v1", "1500"),
				testpbRewardEvent(0, "nom1", "1000"),
				{Section: "Foo", Method: "Foo"},
			},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1500",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect rewards only from events from same transaction",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
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
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "1500",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect rewards to be created for multiple claims",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v1"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "nom1", "1500"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "nom1", "2500"),
				testpbRewardEvent(0, "v1", "3000"),
				testpbRewardEvent(0, "nom1", "3500"),
			},
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
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "2000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "2500",
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
			description:    "expect validator rewards to be created for multiple claims",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v2"}, {102, "v1"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v2", "3000"),
				testpbRewardEvent(0, "v1", "4000"),
			},
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
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "2000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v2",
					StashAccount:          "v2",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "3000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "4000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect validator rewards to be created for multiple claims when there's non reward events",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v2"}, {102, "v1"}},
			events: []*eventpb.Event{
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v1", "1000"),
				{Section: "Foo", Method: "Foo"},
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v1", "2000"),
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v2", "3000"),
				{Section: "Foo", Method: "Foo"},
				testpbRewardEvent(0, "v1", "4000"),
				{Section: "Foo", Method: "Foo"},
			},
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
					EraSequence:           &model.EraSequence{Era: 101},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "2000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v2",
					StashAccount:          "v2",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "3000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 102},
					ValidatorStashAccount: "v1",
					StashAccount:          "v1",
					Kind:                  model.RewardCommissionAndReward,
					Amount:                "4000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect error if event for validator is missing from raw events (all same)",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v1"}, {103, "v1"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v1", "4000"),
			},
			expectErr: true,
		},
		{
			description:    "expect error if event for validator is missing from raw events (one different)",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v2"}, {103, "v1"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "v1", "4000"),
			},
			expectErr: true,
		},
		{
			description:    "expect nominator rewards if  event for validator is missing from raw events but there's only one claim",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events:         []*eventpb.Event{testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectRewards: []model.RewardEraSeq{
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom1",
					Kind:                  model.RewardReward,
					Amount:                "1000",
					Claimed:               true,
				},
				{
					EraSequence:           &model.EraSequence{Era: 100},
					ValidatorStashAccount: "v1",
					StashAccount:          "nom2",
					Kind:                  model.RewardReward,
					Amount:                "2000",
					Claimed:               true,
				},
			},
		},
		{
			description:    "expect error if raw events are in reverse order",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v2"}, {102, "v3"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v3", "1000"),
				testpbRewardEvent(0, "v2", "2000"),
				testpbRewardEvent(0, "v1", "4000"),
			},
			expectErr: true,
		},
		{
			description:    "expect error events are in scrambled order",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}, {101, "v2"}, {102, "v3"}, {102, "v4"}, {102, "v5"}},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "v3", "2000"),
				testpbRewardEvent(0, "v2", "4000"),
				testpbRewardEvent(0, "v4", "2000"),
				testpbRewardEvent(0, "v5", "2000"),
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

			task := NewTransactionParserTask(nil, nil, rewardsMock, syncablesMock, validatorMock)

			rewards, claims, err := task.getRewardsFromEvents(tt.txIdx, tt.rawClaimsForTx, tt.events)
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
		events              []*eventpb.Event
		expectRewards       []model.RewardEraSeq
		expectClaimed       []RewardsClaim
		expectErr           bool
	}{
		{
			description:         "expect claim only for rewards that exist in db",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}, {101, "v1"}, {102, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 0, {101, "v1"}: 2, {102, "v1"}: 0},
			events: []*eventpb.Event{
				testpbRewardEvent(0, "v1", "1000"),
				testpbRewardEvent(0, "nom1", "1500"),
				testpbRewardEvent(0, "v1", "2000"),
				testpbRewardEvent(0, "nom1", "2500"),
				testpbRewardEvent(0, "v1", "3000"),
				testpbRewardEvent(0, "nom1", "3500"),
			},
			expectClaimed: []RewardsClaim{{101, "v1"}},
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
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 1},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectErr:           true,
		},
		{
			description:         "expect error if rewards count > len(rewards)+2",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 5},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectErr:           true,
		},
		{
			description:         "expect claim if rewards count == len(rewards)",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 3},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectClaimed:       []RewardsClaim{{100, "v1"}},
		},
		{
			description:         "expect claim if rewards count == len(rewards)+1",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 4},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
			expectClaimed:       []RewardsClaim{{100, "v1"}},
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

			task := NewTransactionParserTask(nil, nil, rewardsMock, syncablesMock, validatorMock)

			rewards, claims, err := task.getRewardsFromEvents(0, tt.rawClaimsForTx, tt.events)
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
		txIdx                int64
		rawClaimsForTx       []RewardsClaim
		events               []*eventpb.Event
		returnRewardsDbErr   error
		returnSyncablesDbErr error
		expectErr            error
	}{
		{
			description:    "expect no error when db returns no error",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events:         []*eventpb.Event{testpbRewardEvent(0, "v1", "1500")},
		},
		{
			description:        "expect error when rewards db returns error",
			rawClaimsForTx:     []RewardsClaim{{100, "v1"}},
			events:             []*eventpb.Event{testpbRewardEvent(0, "v1", "1500")},
			returnRewardsDbErr: dbErr,
			expectErr:          dbErr,
		},
		{
			description:          "expect error when syncables db returns error",
			rawClaimsForTx:       []RewardsClaim{{100, "v1"}},
			events:               []*eventpb.Event{testpbRewardEvent(0, "v1", "1500")},
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

			task := NewTransactionParserTask(nil, nil, rewardsMock, syncablesMock, validatorMock)

			_, _, err := task.getRewardsFromEvents(tt.txIdx, tt.rawClaimsForTx, tt.events)
			if err != tt.expectErr {
				t.Errorf("Unexpected error on run; got %v, want: %v", err, tt.expectErr)
				return
			}

		})
	}
}

func testpbRewardEvent(txIdx int64, stash, amount string) *eventpb.Event {
	return &eventpb.Event{ExtrinsicIndex: txIdx, Method: "Reward", Section: "staking", Data: []*eventpb.EventData{{Name: "AccountId", Value: stash}, {Name: "Balance", Value: amount}}}
}

func testPayoutStakersTx(stash string, era int64) *transactionpb.Transaction {
	return &transactionpb.Transaction{
		Method:    txMethodPayoutStakers,
		Section:   sectionStaking,
		IsSuccess: true,
		Args:      fmt.Sprintf(`["%v","%v"]`, stash, era),
	}
}
