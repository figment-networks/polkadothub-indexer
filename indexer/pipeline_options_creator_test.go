package indexer

import (
	"errors"
	"testing"

	"github.com/figment-networks/indexing-engine/pipeline"
	mock "github.com/figment-networks/polkadothub-indexer/mock/indexer"
	"github.com/golang/mock/gomock"
)

func TestPipelineOptionsCreator_parse(t *testing.T) {
	expectTasks := []pipeline.TaskName{"Task1", "Task2"}

	t.Run("when version ids are given, return tasks white list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		configParserMock.EXPECT().GetAllTasks(gomock.Any(), gomock.Any()).Return(expectTasks, nil).Times(1)

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

		for _, want := range expectTasks {
			var found bool
			for _, got := range options.TaskWhitelist {
				if got == want {
					found = true
				}
			}
			if !found {
				t.Errorf("expected to find task want: %v; got: %v", want, options.TaskWhitelist)
			}
		}
	})

	t.Run("when target ids are given, return tasks white list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)

		configParserMock.EXPECT().GetAllTasks(gomock.Any(), gomock.Any()).Return(expectTasks, nil).Times(1)

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

		for _, want := range expectTasks {
			var found bool
			for _, got := range options.TaskWhitelist {
				if got == want {
					found = true
				}
			}
			if !found {
				t.Errorf("expected to find task want: %v; got: %v", want, options.TaskWhitelist)
			}
		}
	})

	t.Run("when dry is true, return StagePersistor in StagesBlacklist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configParserMock := mock.NewMockConfigParser(ctrl)
		configParserMock.EXPECT().GetAllTasks(gomock.Any(), gomock.Any()).Return([]pipeline.TaskName{}, nil).Times(1)

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
		configParserMock.EXPECT().GetAllTasks(gomock.Any(), gomock.Any()).Return(nil, testError).Times(1)

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
