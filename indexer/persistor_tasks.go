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
	SyncerPersistorTaskName              = "SyncerPersistor"
	BlockSeqPersistorTaskName            = "BlockSeqPersistor"
	ValidatorSessionSeqPersistorTaskName = "ValidatorSessionSeqPersistor"
	ValidatorEraSeqPersistorTaskName     = "ValidatorEraSeqPersistor"
	ValidatorAggPersistorTaskName        = "ValidatorAggPersistor"
	EventSeqPersistorTaskName            = "EventSeqPersistor"
)

// NewSyncerPersistorTask is responsible for storing syncable to persistence layer
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

// NewBlockSeqPersistorTask is responsible for storing block to persistence layer
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

// NewValidatorSessionSeqPersistorTask is responsible for storing validator session info to persistence layer
func NewValidatorSessionSeqPersistorTask(db *store.Store) pipeline.Task {
	return &validatorSessionSeqPersistorTask{
		db: db,
	}
}

type validatorSessionSeqPersistorTask struct {
	db *store.Store
}

func (t *validatorSessionSeqPersistorTask) GetName() string {
	return ValidatorSessionSeqPersistorTaskName
}

func (t *validatorSessionSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in session [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewValidatorSessionSequences {
		if err := t.db.ValidatorSessionSeq.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorSessionSequences {
		if err := t.db.ValidatorSessionSeq.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewValidatorEraSeqPersistorTask is responsible for storing validator era info to persistence layer
func NewValidatorEraSeqPersistorTask(db *store.Store) pipeline.Task {
	return &validatorEraSeqPersistorTask{
		db: db,
	}
}

type validatorEraSeqPersistorTask struct {
	db *store.Store
}

func (t *validatorEraSeqPersistorTask) GetName() string {
	return ValidatorEraSeqPersistorTaskName
}

func (t *validatorEraSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewValidatorEraSequences {
		if err := t.db.ValidatorEraSeq.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorEraSequences {
		if err := t.db.ValidatorEraSeq.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

func NewValidatorAggPersistorTask(db *store.Store) pipeline.Task {
	return &validatorAggPersistorTask{
		db: db,
	}
}

type validatorAggPersistorTask struct {
	db *store.Store
}

func (t *validatorAggPersistorTask) GetName() string {
	return ValidatorAggPersistorTaskName
}

func (t *validatorAggPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, aggregate := range payload.NewValidatorAggregates {
		if err := t.db.ValidatorAgg.Create(&aggregate); err != nil {
			return err
		}
	}

	for _, aggregate := range payload.UpdatedValidatorAggregates {
		if err := t.db.ValidatorAgg.Save(&aggregate); err != nil {
			return err
		}
	}

	return nil
}

// NewEventSeqPersistorTask is responsible for storing events info to persistence layer
func NewEventSeqPersistorTask(db *store.Store) pipeline.Task {
	return &eventSeqPersistorTask{
		db: db,
	}
}

type eventSeqPersistorTask struct {
	db *store.Store
}

func (t *eventSeqPersistorTask) GetName() string {
	return EventSeqPersistorTaskName
}

func (t *eventSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewEventSequences {
		if err := t.db.EventSeq.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedEventSequences {
		if err := t.db.EventSeq.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}