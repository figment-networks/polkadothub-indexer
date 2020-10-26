package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/pkg/errors"
	"time"
)

var (
	_ pipeline.Sink = (*sink)(nil)
)

func NewSink(db *store.Store, versionNumber int64) *sink {
	return &sink{
		db:            db,
		versionNumber: versionNumber,
	}
}

type sink struct {
	db            *store.Store
	versionNumber int64

	successCount int64
}

func (s *sink) Consume(ctx context.Context, p pipeline.Payload, isForOnlyVersion bool) error {
	payload := p.(*payload)

	logger.DebugJSON(payload,
		logger.Field("process", "pipeline"),
		logger.Field("stage", "sink"),
		logger.Field("height", payload.CurrentHeight),
	)

	if !isForOnlyVersion {
		if err := s.setProcessed(payload.Syncable); err != nil {
			return err
		}

		if err := s.addMetrics(payload); err != nil {
			return err
		}
	} else {
		syncable, err := s.db.Syncables.FindByHeight(payload.CurrentHeight)
		if err != nil {
			return err
		}
		syncable.StartedAt = *types.NewTimeFromTime(time.Now())
		if err := s.setProcessed(syncable); err != nil {
			return err
		}
	}

	s.successCount += 1

	logger.Info(fmt.Sprintf("processing completed [status=success] [height=%d]", payload.CurrentHeight))

	return nil
}

func (s *sink) setProcessed(syncable *model.Syncable) error {
	syncable.MarkProcessed(s.versionNumber)
	if err := s.db.Syncables.Save(syncable); err != nil {
		return errors.Wrap(err, "failed saving syncable in sink")
	}
	return nil
}

func (s *sink) addMetrics(payload *payload) error {
	res, err := s.db.Database.GetTotalSize()
	if err != nil {
		return err
	}

	metric.IndexerHeightSuccess.Inc()
	metric.IndexerHeightDuration.Set(payload.Syncable.Duration.Seconds())
	metric.IndexerDbSizeAfterHeight.Set(res.Size)
	return nil
}
