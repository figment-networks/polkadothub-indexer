package indexer

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	mock_client "github.com/figment-networks/polkadothub-indexer/mock/client"
	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
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

func Test_getClaimedRewardDataFromEvents(t *testing.T) {
	testValidator := "validator_stash1"
	var testEra int64 = 182
	var dbErr = errors.New("test err")

	txtests := []struct {
		description     string
		events          []*eventpb.Event
		validatorEraSeq *model.ValidatorEraSeq
		validatorDbErr  error
		expectErr       error
		expectParsed    parsedRewards
	}{
		{
			description: "expect no rewards if there's no events",
			events:      []*eventpb.Event{},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
		},
		{
			description: "expect StakerRewards  from nominator reward events",
			events:      []*eventpb.Event{testpbRewardEvent(t, "nom1", "2000"), testpbRewardEvent(t, "nom2", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				StakersStake: types.NewQuantityFromInt64(400),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:     true,
				Era:           testEra,
				StakerRewards: []stakerReward{{"nom1", "2000"}, {"nom2", "1200"}},
			},
		},
		{
			description: "expect no StakerRewards  from non-reward events",
			events: []*eventpb.Event{
				testpbRewardEvent(t, "nom1", "2000"),
				&eventpb.Event{Section: sectionStaking, Method: "Foo"},
				&eventpb.Event{Section: "Foo", Method: "Foo"},
			},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				StakersStake: types.NewQuantityFromInt64(400),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:     true,
				Era:           testEra,
				StakerRewards: []stakerReward{{"nom1", "2000"}},
			},
		},
		{
			description: "expect reward from validator when commission is zero",
			events:      []*eventpb.Event{testpbRewardEvent(t, testValidator, "400"), testpbRewardEvent(t, "nom1", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   0,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:     true,
				Era:           testEra,
				Reward:        "400",
				StakerRewards: []stakerReward{{"nom1", "1200"}},
			},
		},
		{
			description: "expect  validator commission when commission = 100%",
			events:      []*eventpb.Event{testpbRewardEvent(t, testValidator, "2000")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   1000000000,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:  true,
				Era:        testEra,
				Commission: "2000",
			},
		},
		{
			description: "expect only validator commission when validator is not staked",
			events:      []*eventpb.Event{testpbRewardEvent(t, testValidator, "2000")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   500000000,
				StakersStake: types.NewQuantityFromInt64(400),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:  true,
				Era:        testEra,
				Commission: "2000",
			},
		},
		{
			description: "expect validator commission_and_reward when validator has commission and is staked",
			events:      []*eventpb.Event{testpbRewardEvent(t, testValidator, "2000"), testpbRewardEvent(t, "nom1", "1200")},
			validatorEraSeq: &model.ValidatorEraSeq{
				Commission:   500000000,
				OwnStake:     types.NewQuantityFromInt64(100),
				StakersStake: types.NewQuantityFromInt64(300),
				TotalStake:   types.NewQuantityFromInt64(400),
			},
			expectParsed: parsedRewards{
				IsClaimed:           true,
				Era:                 testEra,
				RewardAndCommission: "2000",
				StakerRewards:       []stakerReward{{"nom1", "1200"}},
			},
		},
		{
			description: "expect err if db errors",
			events:      []*eventpb.Event{testpbRewardEvent(t, testValidator, "2000"), testpbRewardEvent(t, "nom1", "1200")},
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

			task := NewValidatorsParserTask(nil, nil, nil, syncablesMock, validatorMock)

			got, err := task.getClaimedRewardDataFromEvents(testValidator, testEra, tt.events)
			if err != tt.expectErr {
				t.Errorf("want %v; got %v", tt.expectErr, err)
			}

			if !reflect.DeepEqual(got, tt.expectParsed) {
				t.Errorf("Unexpected parsedReward, want: %v, got: %+v", tt.expectParsed, got)

			}
		})
	}
}

func testpbRewardEvent(t *testing.T, stash, amount string) *eventpb.Event {
	return &eventpb.Event{Method: "Reward", Section: "staking", Data: []*eventpb.EventData{{Name: "AccountId", Value: stash}, {Name: "Balance", Value: amount}}}
}

func testPayoutStakersTx(stash string, era int64) *transactionpb.Annotated {
	return &transactionpb.Annotated{
		Method:  txMethodPayoutStakers,
		Section: sectionStaking,
		Args:    fmt.Sprintf("[\"%v\",\"%v\"]", stash, era),
	}
}
