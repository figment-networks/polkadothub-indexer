package indexer

import (
	"context"
	"fmt"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/pkg/errors"
)

var (
	_ pipeline.Sink = (*sink)(nil)
)

func NewSink(databaseDb store.Database, syncablesDb store.Syncables, versionNumber int64) *sink {
	return &sink{
		databaseDb:    databaseDb,
		syncablesDb:   syncablesDb,
		versionNumber: versionNumber,
	}
}

type sink struct {
	databaseDb    store.Database
	syncablesDb   store.Syncables
	versionNumber int64

	successCount int64
}

func (s *sink) Consume(ctx context.Context, p pipeline.Payload) error {
	payload := p.(*payload)

	logger.DebugJSON(payload,
		logger.Field("process", "pipeline"),
		logger.Field("stage", "sink"),
		logger.Field("height", payload.CurrentHeight),
	)

	if err := s.setProcessed(payload.Syncable); err != nil {
		return err
	}

	if err := s.addMetrics(payload.Syncable); err != nil {
		return err
	}

	s.successCount += 1

	logger.Info(fmt.Sprintf("processing completed [status=success] [height=%d]", payload.CurrentHeight))

	return nil
}

func (s *sink) setProcessed(syncable *model.Syncable) error {
	syncable.MarkProcessed(s.versionNumber)
	if err := s.syncablesDb.SaveSyncable(syncable); err != nil {
		return errors.Wrap(err, "failed saving syncable in sink")
	}
	return nil
}

func (s *sink) addMetrics(syncable *model.Syncable) error {
	res, err := s.databaseDb.GetTotalSize()
	if err != nil {
		return err
	}

	metric.IndexerHeightSuccess.Inc()
	metric.IndexerHeightDuration.Set(syncable.Duration.Seconds())
	metric.IndexerDbSizeAfterHeight.Set(res.Size)
	return nil
}
