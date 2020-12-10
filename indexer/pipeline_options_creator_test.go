package indexer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/figment-networks/indexing-engine/pipeline"
	mock "github.com/figment-networks/polkadothub-indexer/mock/indexer"
	"github.com/golang/mock/gomock"
)

func TestPipelineOptionsCreator_parse(t *testing.T) {
	t.Run("when version ids are given, return tasks white list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		configParserMock.EXPECT().GetTasksByVersionIds(gomock.Any()).Return([]pipeline.TaskName{"task1", "task2"}, nil).Times(1)

		creator := pipelineOptionsCreator{
			configParser: configParserMock,

			desiredVersionIds: []int64{1, 2},
		}

		options, err := creator.parse()
		if err != nil {
			t.Errorf("parse() should not return error")
			return
		}

		if len(options.StagesBlacklist) != 0 {
			t.Errorf("unexpected StagesBlacklist size, want: %d; got: %d", 0, len(options.StagesBlacklist))
		}

		if len(options.TaskWhitelist) != 2 {
			t.Errorf("unexpected TaskWhitelist size, want: %d; got: %d", 2, len(options.TaskWhitelist))
		}

		for i, gotTaskName := range options.TaskWhitelist {
			wantTaskName := fmt.Sprintf("task%d", i+1)

			if string(gotTaskName) != wantTaskName {
				t.Errorf("unexpected task at index %d, want: %s; got: %s", i, wantTaskName, gotTaskName)
			}
		}
	})

	t.Run("when target ids are given, return tasks white list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		configParserMock.EXPECT().GetTasksByTargetIds(gomock.Any()).Return([]pipeline.TaskName{"task1", "task2"}, nil).Times(1)

		creator := pipelineOptionsCreator{
			configParser: configParserMock,

			desiredTargetIds: []int64{1, 2},
		}

		options, err := creator.parse()
		if err != nil {
			t.Errorf("parse() should not return error")
			return
		}

		if len(options.StagesBlacklist) != 0 {
			t.Errorf("unexpected StagesBlacklist size, want: %d; got: %d", 0, len(options.StagesBlacklist))
		}

		if len(options.TaskWhitelist) != 2 {
			t.Errorf("unexpected TaskWhitelist size, want: %d; got: %d", 2, len(options.TaskWhitelist))
		}

		for i, gotTaskName := range options.TaskWhitelist {
			wantTaskName := fmt.Sprintf("task%d", i+1)

			if string(gotTaskName) != wantTaskName {
				t.Errorf("unexpected task at index %d, want: %s; got: %s", i, wantTaskName, gotTaskName)
			}
		}
	})

	t.Run("when dry is true, return StagePersistor in StagesBlacklist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		creator := pipelineOptionsCreator{
			configParser: configParserMock,

			dry: true,
		}

		options, err := creator.parse()
		if err != nil {
			t.Errorf("parse() should not return error")
			return
		}

		if len(options.StagesBlacklist) != 1 {
			t.Errorf("unexpected StagesBlacklist size, want: %d; got: %d", 1, len(options.StagesBlacklist))
			return
		}

		if options.StagesBlacklist[0] != pipeline.StagePersistor {
			t.Errorf("unexpected stage in StagesBlacklist, want: %s; got: %s", pipeline.StagePersistor, options.StagesBlacklist[0])
		}

		if len(options.TaskWhitelist) != 0 {
			t.Errorf("unexpected TaskWhitelist size, want: %d; got: %d", 0, len(options.TaskWhitelist))
		}
	})

	t.Run("when configParser returns error, return that error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		testError := errors.New("test error")
		configParserMock.EXPECT().GetTasksByTargetIds(gomock.Any()).Return(nil, testError).Times(1)

		creator := pipelineOptionsCreator{
			configParser: configParserMock,

			desiredTargetIds: []int64{1, 2},
		}

		options, err := creator.parse()
		if err != testError {
			t.Errorf("parse() should return error %v", testError)
			return
		}

		if options != nil {
			t.Errorf("options should be nil, got: %v", options)
		}
	})
}
