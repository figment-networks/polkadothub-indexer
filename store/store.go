package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
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

type Rewards interface {
	BulkUpsert(records []model.RewardEraSeq) error
	MarkAllClaimed(validatorStash string, era int64, txHash string) error
	GetAll(address string, start, end int64) ([]model.RewardEraSeq, error)
	GetCount(validatorStash string, era int64) (int64, error)
}

type Syncables interface {
	syncables
	FindMostRecenter
}

type SystemEvents interface {
	BulkUpsert(records []model.SystemEvent) error
	FindByActor(actorAddress string, kind *model.SystemEventKind, minHeight *int64) ([]model.SystemEvent, error)
}

type Transactions interface {
	TransactionSeq
}

type Validators interface {
	ValidatorAgg
	ValidatorSeq
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
