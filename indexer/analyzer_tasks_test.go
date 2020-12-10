package indexer

import (
	"encoding/json"
	"errors"
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
	testDelegatorAddress = "test_del_address"
)

var (
	testCfg = &config.Config{
		FirstBlockHeight: 1,
	}
)

func TestSystemEventCreatorTask_getActiveBalanceChangeSystemEvents(t *testing.T) {
	currSyncable := &model.Syncable{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC)),
	}

	currSeq := &model.Sequence{
		Height: currSyncable.Height,
		Time:   currSyncable.Time,
	}

	tests := []struct {
		description             string
		activeBalanceChangeRate float64
		expectedCount           int
		expectedKind            model.SystemEventKind
	}{
		{"returns no system events when active balance haven't changed", 0, 0, ""},
		{"returns no system events when active balance change smaller than 0.1", 0.09, 0, ""},
		{"returns one activeBalanceChange1 system event when active balance change is 0.1", 0.1, 1, model.SystemEventActiveBalanceChange1},
		{"returns one activeBalanceChange1 system events when active balance change is 0.9", 0.9, 1, model.SystemEventActiveBalanceChange1},
		{"returns one activeBalanceChange2 system events when active balance change is 1", 1, 1, model.SystemEventActiveBalanceChange2},
		{"returns one activeBalanceChange2 system events when active balance change is 9", 9, 1, model.SystemEventActiveBalanceChange2},
		{"returns one activeBalanceChange3 system events when active balance change is 10", 10, 1, model.SystemEventActiveBalanceChange3},
		{"returns one activeBalanceChange3 system events when active balance change is 100", 100, 1, model.SystemEventActiveBalanceChange3},
		{"returns one activeBalanceChange3 system events when active balance change is 200", 200, 1, model.SystemEventActiveBalanceChange3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			var activeBalanceBefore int64 = 1000
			activeBalanceAfter := float64(activeBalanceBefore) + (float64(activeBalanceBefore) * tt.activeBalanceChangeRate / 100)

			prevHeightValidatorSequences := []model.ValidatorSeq{
				model.ValidatorSeq{
					Sequence:      currSeq,
					StashAccount:  testValidatorAddress,
					ActiveBalance: types.NewQuantityFromInt64(activeBalanceBefore),
				},
			}
			currHeightValidatorSequences := []model.ValidatorSeq{
				model.ValidatorSeq{
					Sequence:      currSeq,
					StashAccount:  testValidatorAddress,
					ActiveBalance: types.NewQuantityFromInt64(int64(activeBalanceAfter)),
				},
			}

			task := NewSystemEventCreatorTask(testCfg, nil)
			createdSystemEvents, _ := task.getActiveBalanceChangeSystemEvents(currHeightValidatorSequences, prevHeightValidatorSequences, currSyncable)

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

func TestSystemEventCreatorTask_getCommissionChangeSystemEvents(t *testing.T) {
	currSyncable := &model.Syncable{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC)),
	}

	currSeq := &model.EraSequence{
		StartHeight: currSyncable.Height - 10,
		EndHeight:   currSyncable.Height,
		Time:        currSyncable.Time,
	}

	tests := []struct {
		description          string
		commissionChangeRate float64
		expectedCount        int
		expectedKind         model.SystemEventKind
	}{
		{"returns no system events when commission haven't changed", 0, 0, ""},
		{"returns no system events when commission change smaller than 0.1", 0.09, 0, ""},
		{"returns one commissionChange1 system event when commission change is 0.1", 0.1, 1, model.SystemEventCommissionChange1},
		{"returns one commissionChange1 system events when commission change is 0.9", 0.9, 1, model.SystemEventCommissionChange1},
		{"returns one commissionChange2 system events when commission change is 1", 1, 1, model.SystemEventCommissionChange2},
		{"returns one commissionChange2 system events when commission change is 9", 9, 1, model.SystemEventCommissionChange2},
		{"returns one commissionChange3 system events when commission change is 10", 10, 1, model.SystemEventCommissionChange3},
		{"returns one commissionChange3 system events when commission change is 100", 100, 1, model.SystemEventCommissionChange3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			var commissionBefore int64 = 1000
			commissionAfter := float64(commissionBefore) + (float64(commissionBefore) * tt.commissionChangeRate / 100)

			prevHeightValidatorSequences := []model.ValidatorEraSeq{
				model.ValidatorEraSeq{
					EraSequence:  currSeq,
					StashAccount: testValidatorAddress,
					Commission:   commissionBefore,
				},
			}
			currHeightValidatorSequences := []model.ValidatorEraSeq{
				model.ValidatorEraSeq{
					EraSequence:  currSeq,
					StashAccount: testValidatorAddress,
					Commission:   int64(commissionAfter),
				},
			}

			task := NewEraSystemEventCreatorTask(testCfg, nil, nil)
			createdSystemEvents, _ := task.getCommissionChangeSystemEvents(currHeightValidatorSequences, prevHeightValidatorSequences, currSyncable)

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
	currSyncable := &model.Syncable{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC)),
	}

	tests := []struct {
		description   string
		prevSeqs      []model.ValidatorSessionSeq
		currSeqs      []model.ValidatorSessionSeq
		expectedCount int
		expectedKinds []model.SystemEventKind
	}{
		{
			description: "returns no system events when validator is both in prev and current lists",
			prevSeqs: []model.ValidatorSessionSeq{
				{StashAccount: testValidatorAddress},
			},
			currSeqs: []model.ValidatorSessionSeq{
				{StashAccount: testValidatorAddress},
			},
			expectedCount: 0,
		},
		{
			description:   "returns no system events when validator is not in any list",
			prevSeqs:      []model.ValidatorSessionSeq{},
			currSeqs:      []model.ValidatorSessionSeq{},
			expectedCount: 0,
		},
		{
			description:   "returns one joined_set system events when validator is not in prev lists and is in current list",
			prevSeqs:      []model.ValidatorSessionSeq{},
			currSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			expectedCount: 1,
			expectedKinds: []model.SystemEventKind{model.SystemEventJoinedSet},
		},
		{
			description:   "returns one left_set system events when validator is in prev set and not in current set",
			prevSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			currSeqs:      []model.ValidatorSessionSeq{},
			expectedCount: 1,
			expectedKinds: []model.SystemEventKind{model.SystemEventLeftSet},
		},
		{
			description:   "returns 2 joined_set system events when validators are not in prev but are in current lists",
			prevSeqs:      []model.ValidatorSessionSeq{},
			currSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}, {StashAccount: "testValidatorAddress2"}},
			expectedCount: 2,
			expectedKinds: []model.SystemEventKind{model.SystemEventJoinedSet, model.SystemEventJoinedSet},
		},
		{
			description:   "returns 2 left_set system events when validators are in prev but are not in current lists",
			prevSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}, {StashAccount: "testValidatorAddress2"}},
			currSeqs:      []model.ValidatorSessionSeq{},
			expectedCount: 2,
			expectedKinds: []model.SystemEventKind{model.SystemEventLeftSet, model.SystemEventLeftSet},
		},
		{
			description:   "returns left and joined set events",
			prevSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}, {StashAccount: "testValidatorAddress2"}},
			currSeqs:      []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}, {StashAccount: "testValidatorAddress3"}},
			expectedCount: 2,
			expectedKinds: []model.SystemEventKind{model.SystemEventLeftSet, model.SystemEventJoinedSet},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := NewSessionSystemEventCreatorTask(testCfg, nil, nil, nil, nil)
			createdSystemEvents, _ := task.getActiveSetPresenceChangeSystemEvents(tt.currSeqs, tt.prevSeqs, currSyncable)

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

func TestSystemEventCreatorTask_getMissedBlocksSystemEvents(t *testing.T) {
	testSyncable := &model.Syncable{Height: 100, Session: 50}
	var lastSessionHeight int64 = 50
	testErr := errors.New("test err")

	tests := []struct {
		description                string
		missedConsecutiveThreshold int64
		currSeqs                   []model.ValidatorSessionSeq
		prevMissedCounts           map[string]int64
		dbErr                      error

		expectedKinds []model.SystemEventKind
		expectedErr   error
	}{
		{
			description:                "returns no system events when validator is online",
			missedConsecutiveThreshold: 2,
			currSeqs:                   []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress, Online: true}},
			prevMissedCounts:           map[string]int64{testValidatorAddress: 4},
		},
		{
			description:                "returns no system events when validator missed counts are below threshold",
			missedConsecutiveThreshold: 3,
			currSeqs:                   []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			prevMissedCounts:           map[string]int64{testValidatorAddress: 1},
		},
		{
			description:                "returns one system event when validator missed consecutive count equals threshold",
			missedConsecutiveThreshold: 2,
			currSeqs:                   []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			dbErr:                      testErr,
			prevMissedCounts:           map[string]int64{testValidatorAddress: 1},
			expectedErr:                testErr,
		},
		{
			description:                "returns error if db errors",
			missedConsecutiveThreshold: 2,
			currSeqs:                   []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}},
			prevMissedCounts:           map[string]int64{testValidatorAddress: 1},
			expectedKinds:              []model.SystemEventKind{model.SystemEventMissedNConsecutive},
		},
		{
			description:                "returns multiple system events when many validators are offline",
			missedConsecutiveThreshold: 2,
			currSeqs:                   []model.ValidatorSessionSeq{{StashAccount: testValidatorAddress}, {StashAccount: "testValidatorAddress1"}, {StashAccount: "testValidatorAddress2", Online: true}},
			prevMissedCounts:           map[string]int64{testValidatorAddress: 5, "testValidatorAddress1": 5, "testValidatorAddress2": 5},
			expectedKinds:              []model.SystemEventKind{model.SystemEventMissedNConsecutive, model.SystemEventMissedNConsecutive},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			missedConsecutiveThreshold = tt.missedConsecutiveThreshold
			systemEventStoreMock := mock.NewMockSystemEvents(ctrl)

			task := NewSessionSystemEventCreatorTask(testCfg, nil, systemEventStoreMock, nil, nil)

			kind := model.SystemEventMissedNConsecutive
			for _, seq := range tt.currSeqs {
				if seq.Online {
					continue
				}

				dbReturn := []model.SystemEvent{}
				if count, ok := tt.prevMissedCounts[seq.StashAccount]; ok {
					data, err := json.Marshal(model.MissedNConsecutive{Missed: count})
					if err != nil {
						t.Errorf("unexpected error when marshalling data")
						return
					}
					event := model.SystemEvent{Data: types.Jsonb{RawMessage: data}}
					dbReturn = append(dbReturn, event)
				}

				systemEventStoreMock.EXPECT().FindByActor(seq.StashAccount, &kind, &lastSessionHeight).Return(dbReturn, tt.dbErr).Times(1)
				if tt.dbErr != nil {
					break
				}
			}

			createdSystemEvents, err := task.getMissedBlocksSystemEvents(tt.currSeqs, lastSessionHeight, testSyncable)
			if err == nil && tt.expectedErr != nil {
				t.Errorf("should return error")
				return
			}
			if err != nil && tt.expectedErr != err {
				t.Errorf("unexpected error, want %v; got %v", tt.expectedErr, err)
				return
			}

			if len(createdSystemEvents) != len(tt.expectedKinds) {
				t.Errorf("unexpected system event count, want %v; got %v", len(tt.expectedKinds), len(createdSystemEvents))
				return
			}

			for i, kind := range tt.expectedKinds {
				if createdSystemEvents[i].Kind != kind {
					t.Errorf("unexpected system event kind, want %v; got %v", kind, createdSystemEvents[i].Kind)
				}
			}
		})
	}
}

func TestSystemEventCreatorTask_getDelegationChangedSystemEvents(t *testing.T) {
	currSyncable := &model.Syncable{
		Height: 20,
		Time:   *types.NewTimeFromTime(time.Date(2020, 11, 10, 23, 0, 0, 0, time.UTC)),
	}

	tests := []struct {
		description      string
		prevSeqs         []model.AccountEraSeq
		currSeqs         []model.AccountEraSeq
		expectedKinds    []model.SystemEventKind
		expectedAccounts map[string][]string
	}{
		{
			description: "returns no system events when delegator is both in prev and current lists",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},
		},
		{
			description: "returns delegation_joined event when delegator is not in prev but in current lists",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr2", ValidatorStashAccount: testValidatorAddress},
			},
			expectedKinds:    []model.SystemEventKind{model.SystemEventDelegationJoined},
			expectedAccounts: map[string][]string{testValidatorAddress: []string{"addr2"}},
		},
		{
			description: "returns multiple delegation_joined events when delegator is in prev list but not in current",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr1", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
				{StashAccount: "addr1", ValidatorStashAccount: "validatorAddr2"},
			},
			expectedKinds: []model.SystemEventKind{model.SystemEventDelegationJoined, model.SystemEventDelegationJoined},
			expectedAccounts: map[string][]string{
				testValidatorAddress: []string{"addr1"},
				"validatorAddr2":     []string{"addr1"},
			},
		},
		{
			description: "returns delegation_joined event with multiple delegators in data",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr2", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr3", ValidatorStashAccount: testValidatorAddress},
			},
			expectedKinds:    []model.SystemEventKind{model.SystemEventDelegationJoined},
			expectedAccounts: map[string][]string{testValidatorAddress: []string{"addr", "addr2", "addr3"}},
		},
		{
			description: "returns delegation_left event when delegator is in prev list but not in current",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr2", ValidatorStashAccount: testValidatorAddress},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},
			expectedKinds:    []model.SystemEventKind{model.SystemEventDelegationLeft},
			expectedAccounts: map[string][]string{testValidatorAddress: []string{"addr2"}},
		},
		{
			description: "returns delegation_left event with multiple delegators in data",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr2", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr3", ValidatorStashAccount: testValidatorAddress},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
			},

			expectedKinds:    []model.SystemEventKind{model.SystemEventDelegationLeft},
			expectedAccounts: map[string][]string{testValidatorAddress: []string{"addr", "addr2", "addr3"}},
		},
		{
			description: "returns multiple delegation_left events when delegator is in prev list but not in current",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr2", ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
				{StashAccount: "addr2", ValidatorStashAccount: "validatorAddr2"},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
			},
			expectedKinds: []model.SystemEventKind{model.SystemEventDelegationLeft, model.SystemEventDelegationLeft},
			expectedAccounts: map[string][]string{
				testValidatorAddress: []string{"addr2"},
				"validatorAddr2":     []string{"addr2"},
			},
		},
		{
			description: "returns delegation_left and delegation_joined events",
			prevSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
				{StashAccount: "addr2", ValidatorStashAccount: "validatorAddr2"},
			},
			currSeqs: []model.AccountEraSeq{
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: "validatorAddr2"},
				{StashAccount: testDelegatorAddress, ValidatorStashAccount: testValidatorAddress},
				{StashAccount: "addr3", ValidatorStashAccount: testValidatorAddress},
			},
			expectedKinds: []model.SystemEventKind{model.SystemEventDelegationJoined, model.SystemEventDelegationLeft},
			expectedAccounts: map[string][]string{
				testValidatorAddress: []string{"addr3"},
				"validatorAddr2":     []string{"addr2"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := NewEraSystemEventCreatorTask(testCfg, nil, nil)
			createdSystemEvents, _ := task.getDelegationChangedSystemEvents(tt.currSeqs, tt.prevSeqs, currSyncable)

			if len(createdSystemEvents) != len(tt.expectedKinds) {
				t.Errorf("unexpected system event count, want %v; got %v", len(tt.expectedKinds), len(createdSystemEvents))
				return
			}

			for i, event := range createdSystemEvents {
				if event.Kind != tt.expectedKinds[i] {
					t.Errorf("unexpected system event kind, want %v; got %v", tt.expectedKinds[i], event.Kind)
				}

				data := &model.DelegationChangeData{}
				err := json.Unmarshal(event.Data.RawMessage, data)
				if err != nil {
					t.Errorf("unexpected err when unmarshalling data: %v", err)
					return
				}

				if len(data.StashAccounts) != len(tt.expectedAccounts[event.Actor]) {
					t.Errorf("unexpected stash accounts count, want %v; got %v", len(tt.expectedAccounts[event.Actor]), len(data.StashAccounts))
					return
				}

				for _, expected := range tt.expectedAccounts[event.Actor] {
					var found bool
					for _, stash := range data.StashAccounts {
						if stash == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("missing stash account in data, want %v", expected)
					}
				}
			}
		})
	}
}
