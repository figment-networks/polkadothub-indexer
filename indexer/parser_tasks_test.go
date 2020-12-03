package indexer

import (
	"context"
	"testing"

	mock "github.com/figment-networks/polkadothub-indexer/mock/client"
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
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

			mockClient := mock.NewMockAccountClient(ctrl)
			for _, validator := range tt.rawStakingState.GetValidators() {
				mockClient.EXPECT().GetIdentity(validator.StashAccount).Return(&accountpb.GetIdentityResponse{Identity: &accountpb.AccountIdentity{DisplayName: ""}}, nil)
			}

			task := NewValidatorsParserTask(mockClient)

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

	rewardtests := []struct {
		description            string
		rawValidator           *stakingpb.Validator
		totalRewardPoints      int64
		totalRewardPayout      string
		expectCommission       bool
		expectReward           bool
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
			expectNumStakerRewards: 2,
		},
		{description: "does not update ParsedValidators if there's no reward payout",
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
			expectCommission:  true,
		},
	}

	for _, tt := range rewardtests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			mockClient := mock.NewMockAccountClient(ctrl)
			mockClient.EXPECT().GetIdentity(gomock.Any()).Return(nil, nil)

			task := NewValidatorsParserTask(mockClient)

			pl := &payload{
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

			for _, got := range pl.ParsedValidators {

				if tt.expectCommission != (got.UnclaimedCommission != "") {
					t.Errorf("Unexpected UnclaimedCommission, got: %+v", got.UnclaimedCommission)
					return
				}

				if tt.expectReward != (got.UnclaimedReward != "") {
					t.Errorf("Unexpected UnclaimedReward, got: %+v", got.UnclaimedReward)
					return
				}

				if tt.expectNumStakerRewards != len(got.UnclaimedStakerRewards) {
					t.Errorf("Unexpected UnclaimedStakerRewards, want: %v, got: %+v", tt.expectNumStakerRewards, len(got.UnclaimedStakerRewards))
					return
				}
			}
		})
	}
}
