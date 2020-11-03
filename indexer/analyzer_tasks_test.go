package indexer

import (
	"testing"
	"time"

	"github.com/figment-networks/polkadothub-indexer/config"
	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/golang/mock/gomock"
)

const (
	testValidatorAddress = "test_address"
)

var (
	testCfg = &config.Config{
		FirstBlockHeight: 1,
	}
)

func TestSystemEventCreatorTask_getValueChangeSystemEvents(t *testing.T) {
	currSeq := &model.Sequence{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC)),
	}

	tests := []struct {
		description             string
		activeBalanceChangeRate float64
		commissionChangeRate    float64
		expectedCount           int
		expectedKind            model.SystemEventKind
	}{
		{"returns no system events when active balance haven't changed", 0, 0, 0, ""},
		{"returns no system events when active balance change smaller than 0.1", 0.09, 0, 0, ""},
		{"returns one activeBalanceChange1 system event when active balance change is 0.1", 0.1, 0, 1, model.SystemEventActiveBalanceChange1},
		{"returns one activeBalanceChange1 system events when active balance change is 0.9", 0.9, 0, 1, model.SystemEventActiveBalanceChange1},
		{"returns one activeBalanceChange2 system events when active balance change is 1", 1, 0, 1, model.SystemEventActiveBalanceChange2},
		{"returns one activeBalanceChange2 system events when active balance change is 9", 9, 0, 1, model.SystemEventActiveBalanceChange2},
		{"returns one activeBalanceChange3 system events when active balance change is 10", 10, 0, 1, model.SystemEventActiveBalanceChange3},
		{"returns one activeBalanceChange3 system events when active balance change is 100", 100, 0, 1, model.SystemEventActiveBalanceChange3},
		{"returns one activeBalanceChange3 system events when active balance change is 200", 200, 0, 1, model.SystemEventActiveBalanceChange3},

		{"returns no system events when commission haven't changed", 0, 0, 0, ""},
		{"returns no system events when commission change smaller than 0.1", 0, 0.09, 0, ""},
		{"returns one commissionChange1 system event when commission change is 0.1", 0, 0.1, 1, model.SystemEventCommissionChange1},
		{"returns one commissionChange1 system events when commission change is 0.9", 0, 0.9, 1, model.SystemEventCommissionChange1},
		{"returns one commissionChange2 system events when commission change is 1", 0, 1, 1, model.SystemEventCommissionChange2},
		{"returns one commissionChange2 system events when commission change is 9", 0, 9, 1, model.SystemEventCommissionChange2},
		{"returns one commissionChange3 system events when commission change is 10", 0, 10, 1, model.SystemEventCommissionChange3},
		{"returns one commissionChange3 system events when commission change is 100", 0, 100, 1, model.SystemEventCommissionChange3},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validatorSeqStoreMock := mock.NewMockValidatorSeq(ctrl)

			var activeBalanceBefore int64 = 1000
			activeBalanceAfter := float64(activeBalanceBefore) + (float64(activeBalanceBefore) * tt.activeBalanceChangeRate / 100)

			var commissionBefore int64 = 1000
			commissionAfter := float64(commissionBefore) + (float64(commissionBefore) * tt.commissionChangeRate / 100)

			prevHeightValidatorSequences := []model.ValidatorSeq{
				model.ValidatorSeq{
					Sequence:      currSeq,
					StashAccount:  testValidatorAddress,
					ActiveBalance: types.NewQuantityFromInt64(activeBalanceBefore),
					Commission:    types.NewQuantityFromInt64(commissionBefore),
				},
			}
			currHeightValidatorSequences := []model.ValidatorSeq{
				model.ValidatorSeq{
					Sequence:      currSeq,
					StashAccount:  testValidatorAddress,
					ActiveBalance: types.NewQuantityFromInt64(int64(activeBalanceAfter)),
					Commission:    types.NewQuantityFromInt64(int64(commissionAfter)),
				},
			}

			task := NewSystemEventCreatorTask(testCfg, nil, validatorSeqStoreMock, nil)
			createdSystemEvents, _ := task.getValueChangeSystemEvents(currHeightValidatorSequences, prevHeightValidatorSequences)

			if len(createdSystemEvents) != tt.expectedCount {
				t.Errorf("unexpected system event count, want %v; got %v", tt.expectedCount, len(createdSystemEvents))
				return
			}

			if len(createdSystemEvents) > 0 && createdSystemEvents[0].Kind != tt.expectedKind {
				t.Errorf("unexpected system event kind, want %v; got %v", tt.expectedKind, createdSystemEvents[0].Kind)
			}
		})
	}
}

func TestSystemEventCreatorTask_getActiveSetPresenceChangeSystemEvents(t *testing.T) {
	tests := []struct {
		description    string
		prevSeqs       []model.ValidatorSeq // contains waiting and active
		currSeqs       []model.ValidatorSeq // contains waiting and active
		prevActiveSeqs []model.ValidatorSessionSeq
		currActiveSeqs []model.ValidatorSessionSeq
		expectedCount  int
		expectedKinds  []model.SystemEventKind
	}{
		{
			description: "returns no system events when validator is both in all prev and current lists",
			prevSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			prevActiveSeqs: []model.ValidatorSessionSeq{
				{StashAccount: testValidatorAddress},
			},
			currActiveSeqs: []model.ValidatorSessionSeq{
				{StashAccount: testValidatorAddress},
			},
			expectedCount: 0,
		},
		{
			description: "returns no system events when validator is both in prev and current waiting lists",
			prevSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  0,
		},
		{
			description:    "returns no system events when validator is not in any list",
			prevSeqs:       []model.ValidatorSeq{},
			currSeqs:       []model.ValidatorSeq{},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  0,
		},
		{
			description: "returns one joined_waiting_set system events when validator is not in prev lists and is in current list",
			prevSeqs:    []model.ValidatorSeq{},
			currSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  1,
			expectedKinds:  []model.SystemEventKind{model.SystemEventJoinedWaitingSet},
		},
		{
			description: "returns one joined_active_set system events when validator is not in prev active set and is in current active set",
			prevSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{
				{StashAccount: testValidatorAddress},
			},
			expectedCount: 1,
			expectedKinds: []model.SystemEventKind{model.SystemEventJoinedActiveSet},
		},
		{
			description:    "returns one joined_waiting_set system events when validator is in prev active set and not in current active set but still in current",
			prevSeqs:       []model.ValidatorSeq{{StashAccount: testValidatorAddress}},
			currSeqs:       []model.ValidatorSeq{{StashAccount: testValidatorAddress}},
			prevActiveSeqs: []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  1,
			expectedKinds:  []model.SystemEventKind{model.SystemEventJoinedWaitingSet},
		},
		{
			description: "returns one left_set system events when validator is in prev and is not in current lists",
			prevSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs:       []model.ValidatorSeq{},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  1,
			expectedKinds:  []model.SystemEventKind{model.SystemEventLeftSet},
		},
		{
			description:    "returns one left_set system events when validator is in active prev and is not in current lists",
			prevSeqs:       []model.ValidatorSeq{{StashAccount: testValidatorAddress}},
			currSeqs:       []model.ValidatorSeq{},
			prevActiveSeqs: []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  1,
			expectedKinds:  []model.SystemEventKind{model.SystemEventLeftSet},
		},
		{
			description: "returns 2 joined_waiting_set system events when validators are not in prev but are in current lists",
			prevSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs: []model.ValidatorSeq{
				{StashAccount: testValidatorAddress},
				{StashAccount: "addr2"},
				{StashAccount: "addr3"},
			},
			prevActiveSeqs: []model.ValidatorSessionSeq{},
			currActiveSeqs: []model.ValidatorSessionSeq{},
			expectedCount:  2,
			expectedKinds:  []model.SystemEventKind{model.SystemEventJoinedWaitingSet, model.SystemEventJoinedWaitingSet},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := NewSystemEventCreatorTask(testCfg, nil, nil, nil)
			createdSystemEvents, _ := task.getActiveSetPresenceChangeSystemEvents(tt.currSeqs, tt.prevSeqs, tt.currActiveSeqs, tt.prevActiveSeqs, 20)

			if len(createdSystemEvents) != tt.expectedCount {
				t.Errorf("unexpected system event count, want %v; got %v", tt.expectedCount, len(createdSystemEvents))
				return
			}

			for i, kind := range tt.expectedKinds {
				if len(createdSystemEvents) > 0 && createdSystemEvents[i].Kind != kind {
					t.Errorf("unexpected system event kind, want %v; got %v", kind, createdSystemEvents[i].Kind)
				}
			}
		})
	}
}
