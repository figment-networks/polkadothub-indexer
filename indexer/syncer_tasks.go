package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/types"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	MainSyncerTaskName = "MainSyncer"
)

func NewMainSyncerTask(db store.Syncables) pipeline.Task {
	return &mainSyncerTask{
		db: db,
	}
}

type mainSyncerTask struct {
	db store.Syncables
}

func (t *mainSyncerTask) GetName() string {
	return MainSyncerTaskName
}

func (t *mainSyncerTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSyncer, t.GetName(), payload.CurrentHeight))

	syncable, err := t.db.FindByHeight(payload.CurrentHeight)
	if err != nil {
		if err == store.ErrNotFound {
			syncable = &model.Syncable{
				Height: payload.CurrentHeight,
				Time:   payload.HeightMeta.Time,

				ChainUID:      payload.HeightMeta.ChainUID,
				SpecVersion:   payload.HeightMeta.SpecVersion,
				Session:       payload.HeightMeta.Session,
				Era:           payload.HeightMeta.Era,
				LastInEra:     payload.HeightMeta.LastInEra,
				LastInSession: payload.HeightMeta.LastInSession,
				Status:        model.SyncableStatusRunning,
			}
		} else {
			return err
		}
	}

	syncable.StartedAt = *types.NewTimeFromTime(time.Now())

	report, ok := ctx.Value(CtxReport).(*model.Report)
	if ok {
		syncable.ReportID = report.ID
	}

	payload.Syncable = syncable
	return nil
}
