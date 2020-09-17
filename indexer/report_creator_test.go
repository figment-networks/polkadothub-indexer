package indexer

import (
	"testing"
	"time"

	mock "github.com/figment-networks/polkadothub-indexer/mock/store"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestReportCreator_createIfNotExists(t *testing.T) {
	t.Run("where report not found and kinds are the same, should succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		reportStoreMock.EXPECT().FindNotCompletedByIndexVersion(gomock.Any(), gomock.Any()).Return(getTestReport(model.ReportKindSequentialReindex), nil).Times(1)

		creator := reportCreator{
			kind:  model.ReportKindSequentialReindex,
			store: reportStoreMock,
		}

		if err := creator.createIfNotExists(); err != nil {
			t.Errorf("createIfNotExists should not return error, got: %v", err)
			return
		}

		if creator.report == nil {
			t.Errorf("report should not be nil")
		}
	})

	t.Run("when report not found and kinds are different, should return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		reportStoreMock.EXPECT().FindNotCompletedByIndexVersion(gomock.Any(), gomock.Any()).Return(getTestReport(model.ReportKindIndex), nil).Times(1)

		creator := reportCreator{
			kind:  model.ReportKindSequentialReindex,
			store: reportStoreMock,
		}

		if err := creator.createIfNotExists(); err == nil {
			t.Errorf("createIfNotExists should return error")
		}

		if creator.report != nil {
			t.Errorf("report should be nil, got: %v", creator.report)
		}
	})

	t.Run("when FindNotCompletedByIndexVersion returns error, should return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		reportStoreMock.EXPECT().FindNotCompletedByIndexVersion(gomock.Any(), gomock.Any()).Return(nil, errors.New("test error")).Times(1)

		creator := reportCreator{
			kind:  model.ReportKindSequentialReindex,
			store: reportStoreMock,
		}

		if err := creator.createIfNotExists(); err == nil {
			t.Errorf("createIfNotExists should return error")
		}

		if creator.report != nil {
			t.Errorf("report should be nil, got: %v", creator.report)
		}
	})

	t.Run("when Create returns error, should return it", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		testErr := errors.New("test error")
		reportStoreMock.EXPECT().FindNotCompletedByIndexVersion(gomock.Any(), gomock.Any()).Return(nil, store.ErrNotFound).Times(1)
		reportStoreMock.EXPECT().Create(gomock.Any()).Return(testErr).Times(1)

		creator := reportCreator{
			kind:  model.ReportKindSequentialReindex,
			store: reportStoreMock,
		}

		if err := creator.createIfNotExists(); err == nil {
			t.Errorf("createIfNotExists should return error %v", testErr)
		}

		if creator.report != nil {
			t.Errorf("report should be nil, got: %v", creator.report)
		}
	})

	t.Run("when report is found, should assign that report", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		reportStoreMock.EXPECT().FindNotCompletedByIndexVersion(gomock.Any(), gomock.Any()).Return(nil, store.ErrNotFound).Times(1)
		reportStoreMock.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

		creator := reportCreator{
			kind:  model.ReportKindSequentialReindex,
			store: reportStoreMock,
		}

		if err := creator.createIfNotExists(); err != nil {
			t.Errorf("createIfNotExists should not return error, got: %v", err)
			return
		}

		if creator.report == nil {
			t.Errorf("report should not be nil")
		}
	})
}

func TestReportCreator_complete(t *testing.T) {
	t.Run("when no error occurs during save, no error is returned", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		reportStoreMock.EXPECT().Save(gomock.Any()).Return(nil).Times(1)

		creator := reportCreator{
			report: getTestReport(model.ReportKindSequentialReindex),
			store:  reportStoreMock,
		}

		err := creator.complete(10, 10, nil)
		if err != nil {
			t.Errorf("complete() should not return error, got: %v", err)
		}
	})

	t.Run("when error occurs during save, it is returned", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reportStoreMock := mock.NewMockReports(ctrl)

		testErr := errors.New("test error")
		reportStoreMock.EXPECT().Save(gomock.Any()).Return(testErr).Times(1)

		creator := reportCreator{
			report: getTestReport(model.ReportKindSequentialReindex),
			store:  reportStoreMock,
		}

		err := creator.complete(10, 10, nil)
		if err != testErr {
			t.Errorf("complete() should return error %v", testErr)
		}
	})
}

func getTestReport(kind model.ReportKind) *model.Report {
	return &model.Report{
		Model: &model.Model{
			CreatedAt: *types.NewTimeFromTime(time.Now()),
			UpdatedAt: *types.NewTimeFromTime(time.Now()),
		},
		Kind: kind,
	}
}
