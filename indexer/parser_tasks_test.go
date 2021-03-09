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
				Extrinsics: []*blockpb.Extrinsic{
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
				Extrinsics: []*blockpb.Extrinsic{
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
				Extrinsics: []*blockpb.Extrinsic{
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

	markClaimedTest := []struct {
		description   string
		txs           []*transactionpb.Annotated
		expectErr     error
		expectClaimed []RewardsClaim
	}{
		{
			description:   "updates payload if there's a payout stakers transaction",
			txs:           []*transactionpb.Annotated{testPayoutStakersTx(name1, 182)},
			expectClaimed: []RewardsClaim{{182, name1}},
		},
		{
			description:   "updates payload if there's multiple payout stakers transaction",
			txs:           []*transactionpb.Annotated{testPayoutStakersTx(name1, 182), testPayoutStakersTx(name1, 180)},
			expectClaimed: []RewardsClaim{{182, name1}, {180, name1}},
		},
		{
			description: "does not update payload if there's no payout stakers transaction",
			txs:         []*transactionpb.Annotated{{Section: "staking", Method: "Foo"}},
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

			task := NewValidatorsParserTask(nil, nil, rewardsMock, nil, nil)

			pl := &payload{
				RawTransactions: tt.txs,
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
			description:    "expect no rewards if there's no events",
			rawClaimsForTx: []RewardsClaim{{100, "v1"}},
			events:         []*eventpb.Event{},
		},
		{
			description:    "expect no rewards if there's no claims",
			rawClaimsForTx: []RewardsClaim{},
			events:         []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
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

			rewards, err := task.getRewardsFromEvents(tt.txIdx, tt.rawClaimsForTx, tt.events)
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
		expectErr           bool
	}{
		{
			description:         "expect rewards for only claims that don't exist in db",
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
			description:         "expect no rewards and no error if rewards count == len(rewards)",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 3},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
		},
		{
			description:         "expect no rewards and no error if rewards count == len(rewards)+1",
			rawClaimsForTx:      []RewardsClaim{{100, "v1"}},
			returnCountForClaim: map[RewardsClaim]int64{{100, "v1"}: 4},
			events:              []*eventpb.Event{testpbRewardEvent(0, "v1", "1500"), testpbRewardEvent(0, "nom1", "1000"), testpbRewardEvent(0, "nom2", "2000")},
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

			rewards, err := task.getRewardsFromEvents(0, tt.rawClaimsForTx, tt.events)
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
				t.Errorf("Unexpected parsedReward.StakerRewards length, want: %v, got: %+v", len(tt.expectRewards), len(rewards))
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

			_, err := task.getRewardsFromEvents(tt.txIdx, tt.rawClaimsForTx, tt.events)
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

func testPayoutStakersTx(stash string, era int64) *transactionpb.Annotated {
	return &transactionpb.Annotated{
		Method:  txMethodPayoutStakers,
		Section: sectionStaking,
		Args:    fmt.Sprintf("[\"%v\",\"%v\"]", stash, era),
	}
}
