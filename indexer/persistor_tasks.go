package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	SyncerPersistorTaskName              = "SyncerPersistor"
	BlockSeqPersistorTaskName            = "BlockSeqPersistor"
	ValidatorSessionSeqPersistorTaskName = "ValidatorSessionSeqPersistor"
	ValidatorEraSeqPersistorTaskName     = "ValidatorEraSeqPersistor"
	ValidatorAggPersistorTaskName        = "ValidatorAggPersistor"
	EventSeqPersistorTaskName            = "EventSeqPersistor"
	AccountEraSeqPersistorTaskName       = "AccountEraSeqPersistor"
)

// NewSyncerPersistorTask is responsible for storing syncable to persistence layer
func NewSyncerPersistorTask(db store.Syncables) pipeline.Task {
	return &syncerPersistorTask{
		db: db,
	}
}

type syncerPersistorTask struct {
	db store.Syncables
}

func (t *syncerPersistorTask) GetName() string {
	return SyncerPersistorTaskName
}

func (t *syncerPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	return t.db.CreateOrUpdate(payload.Syncable)
}

// NewBlockSeqPersistorTask is responsible for storing block to persistence layer
func NewBlockSeqPersistorTask(db store.BlockSeq) pipeline.Task {
	return &blockSeqPersistorTask{
		db: db,
	}
}

type blockSeqPersistorTask struct {
	db store.BlockSeq
}

func (t *blockSeqPersistorTask) GetName() string {
	return BlockSeqPersistorTaskName
}

func (t *blockSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	if payload.NewBlockSequence != nil {
		return t.db.Create(payload.NewBlockSequence)
	}

	if payload.UpdatedBlockSequence != nil {
		return t.db.Save(payload.UpdatedBlockSequence)
	}

	return nil
}

// NewValidatorSessionSeqPersistorTask is responsible for storing validator session info to persistence layer
func NewValidatorSessionSeqPersistorTask(db store.ValidatorSessionSeq) pipeline.Task {
	return &validatorSessionSeqPersistorTask{
		db: db,
	}
}

type validatorSessionSeqPersistorTask struct {
	db store.ValidatorSessionSeq
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
		if err := t.db.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorSessionSequences {
		if err := t.db.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewValidatorEraSeqPersistorTask is responsible for storing validator era info to persistence layer
func NewValidatorEraSeqPersistorTask(db store.ValidatorEraSeq) pipeline.Task {
	return &validatorEraSeqPersistorTask{
		db: db,
	}
}

type validatorEraSeqPersistorTask struct {
	db store.ValidatorEraSeq
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
		if err := t.db.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorEraSequences {
		if err := t.db.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

func NewValidatorAggPersistorTask(db store.ValidatorAgg) pipeline.Task {
	return &validatorAggPersistorTask{
		db: db,
	}
}

type validatorAggPersistorTask struct {
	db store.ValidatorAgg
}

func (t *validatorAggPersistorTask) GetName() string {
	return ValidatorAggPersistorTaskName
}

func (t *validatorAggPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, aggregate := range payload.NewValidatorAggregates {
		if err := t.db.Create(&aggregate); err != nil {
			return err
		}
	}

	for _, aggregate := range payload.UpdatedValidatorAggregates {
		if err := t.db.Save(&aggregate); err != nil {
			return err
		}
	}

	return nil
}

// NewEventSeqPersistorTask is responsible for storing events info to persistence layer
func NewEventSeqPersistorTask(db store.EventSeq) pipeline.Task {
	return &eventSeqPersistorTask{
		db: db,
	}
}

type eventSeqPersistorTask struct {
	db store.EventSeq
}

func (t *eventSeqPersistorTask) GetName() string {
	return EventSeqPersistorTaskName
}

func (t *eventSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewEventSequences {
		if err := t.db.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedEventSequences {
		if err := t.db.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewAccountEraSeqPersistorTask is responsible for storing account era info to persistence layer
func NewAccountEraSeqPersistorTask(db store.AccountEraSeq) pipeline.Task {
	return &accountEraSeqPersistorTask{
		db: db,
	}
}

type accountEraSeqPersistorTask struct {
	db store.AccountEraSeq
}

func (t *accountEraSeqPersistorTask) GetName() string {
	return AccountEraSeqPersistorTaskName
}

func (t *accountEraSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewAccountEraSequences {
		if err := t.db.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedAccountEraSequences {
		if err := t.db.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}
