package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Accounts interface {
	AccountEraSeq
}

type Blocks interface {
	BlockSeq
	BlockSummary
}

type Database interface {
	GetTotalSize() (*GetTotalSizeResult, error)
}

type Events interface {
	EventSeq
}

type Reports interface {
	baseStore
	DeleteByKinds(kinds []model.ReportKind) error
	FindNotCompletedByIndexVersion(indexVersion int64, kinds ...model.ReportKind) (*model.Report, error)
	FindNotCompletedByKind(kinds ...model.ReportKind) (*model.Report, error)
	Last() (*model.Report, error)
}

type Syncables interface {
	CreateOrUpdate(val *model.Syncable) error
	FindByHeight(height int64) (syncable *model.Syncable, err error)
	FindFirstByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindLastInEra(era int64) (syncable *model.Syncable, err error)
	FindLastInSession(session int64) (syncable *model.Syncable, err error)
	FindMostRecent() (*model.Syncable, error)
	FindMostRecentByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindSmallestIndexVersion() (*int64, error)
	SaveSyncable(*model.Syncable) error
	SetProcessedAtForRange(reportID types.ID, startHeight int64, endHeight int64) error
}

type Validators interface {
	ValidatorAgg
	ValidatorEraSeq
	ValidatorSessionSeq
	ValidatorSummary
}

type GetTotalSizeResult struct {
	Size float64 `json:"size"`
}

type baseStore interface {
	Create(record interface{}) error
	Update(record interface{}) error
	Save(record interface{}) error
}
