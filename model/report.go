package model

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"time"
)

const (
	ReportKindIndex ReportKind = iota + 1
	ReportKindParallelReindex
	ReportKindSequentialReindex
)

type Report struct {
	*Model

	Kind         ReportKind
	IndexVersion int64
	StartHeight  int64
	EndHeight    int64
	SuccessCount *int64
	ErrorCount   *int64
	ErrorMsg     *string
	Duration     time.Duration
	CompletedAt  *types.Time
}

type ReportKind int

func (k ReportKind) String() string {
	switch k {
	case ReportKindIndex:
		return "index"
	case ReportKindParallelReindex:
		return "parallel_reindex"
	case ReportKindSequentialReindex:
		return "sequential_reindex"
	default:
		return "unknown"
	}
}

func (Report) TableName() string {
	return "reports"
}

func (r *Report) Valid() bool {
	return r.StartHeight >= 0 &&
		r.EndHeight >= 0
}

func (r *Report) Equal(m Report) bool {
	return m.Model != nil &&
		r.Model.ID == m.Model.ID
}

func (r *Report) Complete(successCount int64, errorCount int64, err error) {
	completedAt := types.NewTimeFromTime(time.Now())

	r.SuccessCount = &successCount
	r.ErrorCount = &errorCount
	r.Duration = time.Since(r.CreatedAt.Time)
	r.CompletedAt = completedAt

	if err != nil {
		errMsg := err.Error()
		r.ErrorMsg = &errMsg
	}
}
