package indexer

import (
	"context"
	"reflect"
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestValidatorAggCreatorTask_Run(t *testing.T) {
	syncTime := *types.NewTimeFromTime(time.Now())
	const syncHeight int64 = 31
	dbErr := errors.New("unexpected err")

	tests := []struct {
		description      string
		parsedValidators ParsedValidatorsData
		syncable         model.Syncable
		expectErr        error
		expectValidators []model.ValidatorAgg
	}{
		{
			description: "Adds new validator to payload.NewValidatorAggregates",
			parsedValidators: map[string]parsedValidator{
				"stashAcct1": {
					Performance: &validatorperformancepb.Validator{
						Online: true,
					},
				},
			},
			syncable: model.Syncable{
				Height:        syncHeight,
				Time:          syncTime,
				LastInSession: false,
			},
			expectErr: nil,
			expectValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: syncHeight,
					StartedAt:       syncTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: syncHeight,
				},
			},
		},
		{
			description: "Adds Uptime data to new validators if block is last in session",
			parsedValidators: map[string]parsedValidator{
				"stashAcct1": {
					Performance: &validatorperformancepb.Validator{
						Online: true,
					},
				},
				"stashAcct2": {
					Performance: &validatorperformancepb.Validator{
						Online: false,
					},
				},
			},
			syncable: model.Syncable{
				Height:        syncHeight,
				Time:          syncTime,
				LastInSession: true,
			},
			expectErr: nil,
			expectValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: syncHeight,
					StartedAt:       syncTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: syncHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  1,
				},
				{Aggregate: &model.Aggregate{
					StartedAtHeight: syncHeight,
					StartedAt:       syncTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct2",
					RecentAsValidatorHeight: syncHeight,
					AccumulatedUptime:       0,
					AccumulatedUptimeCount:  1,
				},
			},
		},
		{
			description: "Returns err on unexpected dberr",
			parsedValidators: map[string]parsedValidator{
				"stashAcct1": {
					Performance: &validatorperformancepb.Validator{
						Online: true,
					},
				},
			},
			syncable: model.Syncable{
				Height:        syncHeight,
				Time:          syncTime,
				LastInSession: false,
			},
			expectErr:        dbErr,
			expectValidators: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // need to set this since running tests in parallel
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorAgg(ctrl)

			pld := &payload{
				ParsedValidators: tt.parsedValidators,
				Syncable:         &tt.syncable,
			}

			for key := range tt.parsedValidators {
				if tt.expectErr == dbErr {
					dbMock.EXPECT().FindAggByStashAccount(key).Return(nil, dbErr).Times(1)
					break
				}
				dbMock.EXPECT().FindAggByStashAccount(key).Return(nil, store.ErrNotFound).Times(1)
			}

			task := NewValidatorAggCreatorTask(dbMock)
			if err := task.Run(ctx, pld); err != tt.expectErr {
				t.Errorf("unexpected error, got: %v; want: %v", err, tt.expectErr)
				return
			}

			// don't check payload if expected error
			if tt.expectErr != nil {
				return
			}

			if len(pld.NewValidatorAggregates) != len(tt.expectValidators) {
				t.Errorf("expected payload.NewValidatorAggregates to contain new accounts, got: %v; want: %v", len(pld.NewValidatorAggregates), len(tt.expectValidators))
				return
			}

			for _, expected := range tt.expectValidators {
				var found bool
				for _, got := range pld.NewValidatorAggregates {
					if got.StashAccount == expected.StashAccount {
						if !reflect.DeepEqual(got, expected) {
							t.Errorf("unexpected entry in payload.NewAggregatedValidators, got: %v; want: %v", got, expected)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing entry in payload.NewAggregatedValidators, want: %v", expected)
				}
			}

		})
	}

	startedAtTime := *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC))
	const startedAtHeight int64 = 30

	updateValidatorTests := []struct {
		description      string
		parsedValidators ParsedValidatorsData
		returnValidators []model.ValidatorAgg
		syncable         model.Syncable
		expectValidators []model.ValidatorAgg
	}{
		{
			description: "Adds validator to payload.UpdatedValidatorAggregates",
			parsedValidators: map[string]parsedValidator{
				"stashAcct1": {
					Performance: nil,
				},
			},
			returnValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  startedAtHeight,
					RecentAt:        startedAtTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: startedAtHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  1,
				},
			},
			syncable: model.Syncable{
				Height:        syncHeight,
				Time:          syncTime,
				LastInSession: false,
			},
			expectValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: syncHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  1,
				},
			},
		},
		{
			description: "Adds Uptime data to existing validators if block is last in session",
			parsedValidators: map[string]parsedValidator{
				"stashAcct1": {
					Performance: &validatorperformancepb.Validator{
						Online: true,
					},
				},
				"stashAcct2": {
					Performance: &validatorperformancepb.Validator{
						Online: false,
					},
				},
			},
			returnValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  startedAtHeight,
					RecentAt:        startedAtTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: startedAtHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  2,
				},
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  startedAtHeight,
					RecentAt:        startedAtTime,
				},
					StashAccount:            "stashAcct2",
					RecentAsValidatorHeight: startedAtHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  2,
				},
			},
			syncable: model.Syncable{
				Height:        syncHeight,
				Time:          syncTime,
				LastInSession: true,
			},
			expectValidators: []model.ValidatorAgg{
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct1",
					RecentAsValidatorHeight: syncHeight,
					AccumulatedUptime:       2,
					AccumulatedUptimeCount:  3,
				},
				{Aggregate: &model.Aggregate{
					StartedAtHeight: startedAtHeight,
					StartedAt:       startedAtTime,
					RecentAtHeight:  syncHeight,
					RecentAt:        syncTime,
				},
					StashAccount:            "stashAcct2",
					RecentAsValidatorHeight: syncHeight,
					AccumulatedUptime:       1,
					AccumulatedUptimeCount:  3,
				},
			},
		},
	}

	for _, tt := range updateValidatorTests {
		tt := tt // need to set this since running tests in parallel
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			ctx := context.Background()

			dbMock := mock.NewMockValidatorAgg(ctrl)

			pld := &payload{
				ParsedValidators: tt.parsedValidators,
				Syncable:         &tt.syncable,
			}

			for _, validator := range tt.returnValidators {
				expect := validator
				dbMock.EXPECT().FindAggByStashAccount(validator.StashAccount).Return(&expect, nil).Times(1)
			}

			task := NewValidatorAggCreatorTask(dbMock)
			if err := task.Run(ctx, pld); err != nil {
				t.Errorf("unexpected error, got: %v", err)
				return
			}

			if len(pld.UpdatedValidatorAggregates) != len(tt.expectValidators) {
				t.Errorf("expected payload.UpdatedValidatorAggregates to contain accounts, got: %v; want: %v", len(pld.UpdatedValidatorAggregates), len(tt.expectValidators))
				return
			}

			for _, expected := range tt.expectValidators {
				var found bool
				for _, got := range pld.UpdatedValidatorAggregates {
					if got.StashAccount == expected.StashAccount {
						if !reflect.DeepEqual(got, expected) {
							t.Errorf("unexpected entry in payload.UpdatedValidatorAggregates, got: %v; want: %v", got, expected)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing entry in payload.UpdatedValidatorAggregates, want: %v", expected)
				}
			}

		})
	}
}
