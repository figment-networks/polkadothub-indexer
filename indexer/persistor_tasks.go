package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"time"
)

const (
	SyncerPersistorTaskName       = "SyncerPersistor"
	BlockSeqPersistorTaskName     = "BlockSeqPersistor"
	ValidatorSeqPersistorTaskName = "ValidatorSeqPersistor"
	ValidatorAggPersistorTaskName = "ValidatorAggPersistor"
)

func NewSyncerPersistorTask(db *store.Store) pipeline.Task {
	return &syncerPersistorTask{
		db: db,
	}
}

type syncerPersistorTask struct {
	db *store.Store
}

func (t *syncerPersistorTask) GetName() string {
	return SyncerPersistorTaskName
}

func (t *syncerPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	return t.db.Syncables.CreateOrUpdate(payload.Syncable)
}

func NewBlockSeqPersistorTask(db *store.Store) pipeline.Task {
	return &blockSeqPersistorTask{
		db: db,
	}
}

type blockSeqPersistorTask struct {
	db *store.Store
}

func (t *blockSeqPersistorTask) GetName() string {
	return BlockSeqPersistorTaskName
}

func (t *blockSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	if payload.NewBlockSequence != nil {
		return t.db.BlockSeq.Create(payload.NewBlockSequence)
	}

	if payload.UpdatedBlockSequence != nil {
		return t.db.BlockSeq.Save(payload.UpdatedBlockSequence)
	}

	return nil
}

//func NewValidatorSeqPersistorTask(db *store.Store) pipeline.Task {
//	return &validatorSeqPersistorTask{
//		db: db,
//	}
//}
//
//type validatorSeqPersistorTask struct {
//	db *store.Store
//}
//
//func (t *validatorSeqPersistorTask) GetName() string {
//	return ValidatorSeqPersistorTaskName
//}
//
//func (t *validatorSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))
//
//	for _, sequence := range payload.NewValidatorSequences {
//		if err := t.db.ValidatorSeq.Create(&sequence); err != nil {
//			return err
//		}
//	}
//
//	for _, sequence := range payload.UpdatedValidatorSequences {
//		if err := t.db.ValidatorSeq.Save(&sequence); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func NewValidatorAggPersistorTask(db *store.Store) pipeline.Task {
//	return &validatorAggPersistorTask{
//		db: db,
//	}
//}
//
//type validatorAggPersistorTask struct {
//	db *store.Store
//}
//
//func (t *validatorAggPersistorTask) GetName() string {
//	return ValidatorAggPersistorTaskName
//}
//
//func (t *validatorAggPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))
//
//	for _, aggregate := range payload.NewAggregatedValidators {
//		if err := t.db.ValidatorAgg.Create(&aggregate); err != nil {
//			return err
//		}
//	}
//
//	for _, aggregate := range payload.UpdatedAggregatedValidators {
//		if err := t.db.ValidatorAgg.Save(&aggregate); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
