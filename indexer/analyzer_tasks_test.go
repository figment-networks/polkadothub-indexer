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
		t.Run(tt.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validatorSeqStoreMock := mock.NewMockValidatorSeq(ctrl)

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

			task := NewSystemEventCreatorTask(testCfg, validatorSeqStoreMock)
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
