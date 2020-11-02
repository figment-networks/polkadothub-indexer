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
	TransactionSeqPersistorTaskName      = "TransactionSeqPersistor"
	ValidatorSeqPersistorTaskName        = "ValidatorSeqPersistor"
	SystemEventPersistorTaskName         = "SystemEventPersistor"
)

// NewSyncerPersistorTask is responsible for storing syncable to persistence layer
func NewSyncerPersistorTask(syncablesDb store.Syncables) pipeline.Task {
	return &syncerPersistorTask{
		syncablesDb: syncablesDb,
	}
}

type syncerPersistorTask struct {
	syncablesDb store.Syncables
}

func (t *syncerPersistorTask) GetName() string {
	return SyncerPersistorTaskName
}

func (t *syncerPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	return t.syncablesDb.CreateOrUpdate(payload.Syncable)
}

// NewBlockSeqPersistorTask is responsible for storing block to persistence layer
func NewBlockSeqPersistorTask(blockSeqDb store.BlockSeq) pipeline.Task {
	return &blockSeqPersistorTask{
		blockSeqDb: blockSeqDb,
	}
}

type blockSeqPersistorTask struct {
	blockSeqDb store.BlockSeq
}

func (t *blockSeqPersistorTask) GetName() string {
	return BlockSeqPersistorTaskName
}

func (t *blockSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	if payload.NewBlockSequence != nil {
		return t.blockSeqDb.CreateSeq(payload.NewBlockSequence)
	}

	if payload.UpdatedBlockSequence != nil {
		return t.blockSeqDb.SaveSeq(payload.UpdatedBlockSequence)
	}

	return nil
}

// NewValidatorSessionSeqPersistorTask is responsible for storing validator session info to persistence layer
func NewValidatorSessionSeqPersistorTask(validatorSessionSeqDb store.ValidatorSessionSeq) pipeline.Task {
	return &validatorSessionSeqPersistorTask{
		validatorSessionSeqDb: validatorSessionSeqDb,
	}
}

type validatorSessionSeqPersistorTask struct {
	validatorSessionSeqDb store.ValidatorSessionSeq
}

func (t *validatorSessionSeqPersistorTask) GetName() string {
	return ValidatorSessionSeqPersistorTaskName
}

func (t *validatorSessionSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in session [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewValidatorSessionSequences {
		if err := t.validatorSessionSeqDb.CreateSessionSeq(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorSessionSequences {
		if err := t.validatorSessionSeqDb.SaveSessionSeq(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewValidatorEraSeqPersistorTask is responsible for storing validator era info to persistence layer
func NewValidatorEraSeqPersistorTask(validatorEraSeqDb store.ValidatorEraSeq) pipeline.Task {
	return &validatorEraSeqPersistorTask{
		validatorEraSeqDb: validatorEraSeqDb,
	}
}

type validatorEraSeqPersistorTask struct {
	validatorEraSeqDb store.ValidatorEraSeq
}

func (t *validatorEraSeqPersistorTask) GetName() string {
	return ValidatorEraSeqPersistorTaskName
}

func (t *validatorEraSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewValidatorEraSequences {
		if err := t.validatorEraSeqDb.CreateEraSeq(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorEraSequences {
		if err := t.validatorEraSeqDb.SaveEraSeq(&sequence); err != nil {
			return err
		}
	}

	return nil
}

func NewValidatorAggPersistorTask(validatorAggDb store.ValidatorAgg) pipeline.Task {
	return &validatorAggPersistorTask{
		validatorAggDb: validatorAggDb,
	}
}

type validatorAggPersistorTask struct {
	validatorAggDb store.ValidatorAgg
}

func (t *validatorAggPersistorTask) GetName() string {
	return ValidatorAggPersistorTaskName
}

func (t *validatorAggPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, aggregate := range payload.NewValidatorAggregates {
		if err := t.validatorAggDb.CreateAgg(&aggregate); err != nil {
			return err
		}
	}

	for _, aggregate := range payload.UpdatedValidatorAggregates {
		if err := t.validatorAggDb.SaveAgg(&aggregate); err != nil {
			return err
		}
	}

	return nil
}

// NewEventSeqPersistorTask is responsible for storing events info to persistence layer
func NewEventSeqPersistorTask(eventSeqDb store.EventSeq) pipeline.Task {
	return &eventSeqPersistorTask{
		eventSeqDb: eventSeqDb,
	}
}

type eventSeqPersistorTask struct {
	eventSeqDb store.EventSeq
}

func (t *eventSeqPersistorTask) GetName() string {
	return EventSeqPersistorTaskName
}

func (t *eventSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewEventSequences {
		if err := t.eventSeqDb.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedEventSequences {
		if err := t.eventSeqDb.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewAccountEraSeqPersistorTask is responsible for storing account era info to persistence layer
func NewAccountEraSeqPersistorTask(accountEraSeqDb store.AccountEraSeq) pipeline.Task {
	return &accountEraSeqPersistorTask{
		accountEraSeqDb: accountEraSeqDb,
	}
}

type accountEraSeqPersistorTask struct {
	accountEraSeqDb store.AccountEraSeq
}

func (t *accountEraSeqPersistorTask) GetName() string {
	return AccountEraSeqPersistorTaskName
}

func (t *accountEraSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewAccountEraSequences {
		if err := t.accountEraSeqDb.Create(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedAccountEraSequences {
		if err := t.accountEraSeqDb.Save(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewTransactionSeqPersistorTask is responsible for storing transaction info to persistence layer
func NewTransactionSeqPersistorTask(transactionSeqDb store.TransactionSeq) pipeline.Task {
	return &transactionSeqPersistorTask{
		transactionSeqDb: transactionSeqDb,
	}
}

type transactionSeqPersistorTask struct {
	transactionSeqDb store.TransactionSeq
}

func (t *transactionSeqPersistorTask) GetName() string {
	return TransactionSeqPersistorTaskName
}

func (t *transactionSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload, ok := p.(*payload)
	if !ok {
		return fmt.Errorf("Interface is not a  *payload type (%T)", p)
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewTransactionSequences {
		if err := t.transactionSeqDb.CreateTransactionSeq(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedTransactionSequences {
		if err := t.transactionSeqDb.SaveTransactionSeq(&sequence); err != nil {
			return err
		}
	}

	return nil
}

// NewValidatorSeqPersistorTask is responsible for storing transaction info to persistence layer
func NewValidatorSeqPersistorTask(ValidatorSeqDb store.ValidatorSeq) pipeline.Task {
	return &validatorSeqPersistorTask{
		ValidatorSeqDb: ValidatorSeqDb,
	}
}

type validatorSeqPersistorTask struct {
	ValidatorSeqDb store.ValidatorSeq
}

func (t *validatorSeqPersistorTask) GetName() string {
	return ValidatorSeqPersistorTaskName
}

func (t *validatorSeqPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload, ok := p.(*payload)
	if !ok {
		return fmt.Errorf("Interface is not a  *payload type (%T)", p)
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, sequence := range payload.NewValidatorSequences {
		if err := t.ValidatorSeqDb.CreateSeq(&sequence); err != nil {
			return err
		}
	}

	for _, sequence := range payload.UpdatedValidatorSequences {
		if err := t.ValidatorSeqDb.SaveSeq(&sequence); err != nil {
			return err
		}
	}

	return nil
}

func NewSystemEventPersistorTask(systemEventDb store.SystemEvents) pipeline.Task {
	return &systemEventPersistorTask{
		systemEventDb: systemEventDb,
	}
}

type systemEventPersistorTask struct {
	systemEventDb store.SystemEvents
}

func (t *systemEventPersistorTask) GetName() string {
	return SystemEventPersistorTaskName
}

func (t *systemEventPersistorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StagePersistor, t.GetName(), payload.CurrentHeight))

	for _, systemEvent := range payload.SystemEvents {
		if err := t.systemEventDb.CreateOrUpdate(systemEvent); err != nil {
			return err
		}
	}

	return nil
}
