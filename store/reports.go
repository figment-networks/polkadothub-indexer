package store

import "github.com/figment-networks/polkadothub-indexer/model"

type Reports interface {
	BaseStore
	DeleteByKinds(kinds []model.ReportKind) error
	FindNotCompletedByIndexVersion(indexVersion int64, kinds ...model.ReportKind) (*model.Report, error)
	FindNotCompletedByKind(kinds ...model.ReportKind) (*model.Report, error)
	Last() (*model.Report, error)
}
