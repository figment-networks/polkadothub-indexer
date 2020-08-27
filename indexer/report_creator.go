package indexer

import (
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

// reportCreator creates and completes report
type reportCreator struct {
	kind         model.ReportKind
	indexVersion int64
	startHeight  int64
	endHeight    int64

	store ReportStore

	report *model.Report
}

type ReportStore interface {
	FindNotCompletedByIndexVersion(int64, ...model.ReportKind) (*model.Report, error)
	Create(interface{}) error
	Save(interface{}) error
}

func (o *reportCreator) createIfNotExists(kinds ...model.ReportKind) error {
	report, err := o.store.FindNotCompletedByIndexVersion(o.indexVersion, kinds...)
	if err != nil {
		if err == store.ErrNotFound {
			if err = o.create(); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if report.Kind != o.kind {
			return errors.New(fmt.Sprintf("there is already reindexing in process [kind=%s] (use -force flag to override it)", report.Kind))
		}
		o.report = report
	}
	return nil
}

func (o *reportCreator) create() error {
	report := &model.Report{
		Kind:         o.kind,
		IndexVersion: o.indexVersion,
		StartHeight:  o.startHeight,
		EndHeight:    o.endHeight,
	}

	if err := o.store.Create(report); err != nil {
		return err
	}

	o.report = report

	return nil
}

func (o *reportCreator) complete(totalCount int64, successCount int64, err error) error {
	o.report.Complete(successCount, totalCount-successCount, err)

	return o.store.Save(o.report)
}
