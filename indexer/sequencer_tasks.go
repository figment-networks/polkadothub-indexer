package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"time"
)

const (
	BlockSeqCreatorTaskName            = "BlockSeqCreator"
	ValidatorSessionSeqCreatorTaskName = "ValidatorSessionSeqCreator"
	ValidatorEraSeqCreatorTaskName     = "ValidatorEraSeqCreator"
	EventSeqCreatorTaskName            = "EventSeqCreator"
)

var (
	_ pipeline.Task = (*blockSeqCreatorTask)(nil)
	_ pipeline.Task = (*validatorSessionSeqCreatorTask)(nil)
)

// NewBlockSeqCreatorTask creates block sequences
func NewBlockSeqCreatorTask(db *store.Store) *blockSeqCreatorTask {
	return &blockSeqCreatorTask{
		db: db,
	}
}

type blockSeqCreatorTask struct {
	db *store.Store
}

func (t *blockSeqCreatorTask) GetName() string {
	return BlockSeqCreatorTaskName
}

func (t *blockSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	mappedBlockSeq, err := ToBlockSequence(payload.Syncable, payload.RawBlock, payload.ParsedBlock)
	if err != nil {
		return err
	}

	blockSeq, err := t.db.BlockSeq.FindByHeight(payload.CurrentHeight)
	if err != nil {
		if err == store.ErrNotFound {
			payload.NewBlockSequence = mappedBlockSeq
			return nil
		} else {
			return err
		}
	}

	blockSeq.Update(*mappedBlockSeq)
	payload.UpdatedBlockSequence = blockSeq

	return nil
}

// NewValidatorSessionSeqCreatorTask creates validator session sequences
func NewValidatorSessionSeqCreatorTask(cfg *config.Config, db *store.Store) *validatorSessionSeqCreatorTask {
	return &validatorSessionSeqCreatorTask{
		cfg: cfg,
		db:  db,
	}
}

type validatorSessionSeqCreatorTask struct {
	cfg *config.Config
	db  *store.Store
}

func (t *validatorSessionSeqCreatorTask) GetName() string {
	return ValidatorSessionSeqCreatorTaskName
}

func (t *validatorSessionSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in session [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	var firstHeightInSession int64
	lastSyncableInPrevSession, err := t.db.Syncables.FindLastInSession(payload.Syncable.Session - 1)
	if err != nil {
		if err == store.ErrNotFound {
			firstHeightInSession = t.cfg.FirstBlockHeight
		} else {
			return err
		}
	} else {
		firstHeightInSession = lastSyncableInPrevSession.Height + 1
	}

	mappedValidatorSessionSeqs, err := ToValidatorSessionSequence(payload.Syncable, firstHeightInSession, payload.RawValidatorPerformance)
	if err != nil {
		return err
	}

	var newValidatorSessionSeqs []model.ValidatorSessionSeq
	var updatedValidatorSessionSeqs []model.ValidatorSessionSeq
	for _, rawValidatorSessionSeq := range mappedValidatorSessionSeqs {
		validatorSessionSeq, err := t.db.ValidatorSessionSeq.FindBySessionAndStashAccount(payload.Syncable.Session, rawValidatorSessionSeq.StashAccount)
		if err != nil {
			if err == store.ErrNotFound {
				newValidatorSessionSeqs = append(newValidatorSessionSeqs, rawValidatorSessionSeq)
				continue
			} else {
				return err
			}
		}

		validatorSessionSeq.Update(rawValidatorSessionSeq)
		updatedValidatorSessionSeqs = append(updatedValidatorSessionSeqs, *validatorSessionSeq)
	}

	payload.NewValidatorSessionSequences = newValidatorSessionSeqs
	payload.UpdatedValidatorSessionSequences = updatedValidatorSessionSeqs

	return nil
}

// NewValidatorEraSeqCreatorTask creates validator era sequences
func NewValidatorEraSeqCreatorTask(cfg *config.Config, db *store.Store) *validatorEraSeqCreatorTask {
	return &validatorEraSeqCreatorTask{
		cfg: cfg,
		db:  db,
	}
}

type validatorEraSeqCreatorTask struct {
	cfg *config.Config
	db  *store.Store
}

func (t *validatorEraSeqCreatorTask) GetName() string {
	return ValidatorEraSeqCreatorTaskName
}

func (t *validatorEraSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.db.Syncables.FindLastInEra(payload.Syncable.Era - 1)
	if err != nil {
		if err == store.ErrNotFound {
			firstHeightInEra = t.cfg.FirstBlockHeight
		} else {
			return err
		}
	} else {
		firstHeightInEra = lastSyncableInPrevEra.Height + 1
	}

	mappedValidatorEraSeqs, err := ToValidatorEraSequence(payload.Syncable, firstHeightInEra, payload.RawStaking.GetValidators())
	if err != nil {
		return err
	}

	var newValidatorEraSeqs []model.ValidatorEraSeq
	var updatedValidatorEraSeqs []model.ValidatorEraSeq
	for _, rawValidatorEraSeq := range mappedValidatorEraSeqs {
		validatorEraSeq, err := t.db.ValidatorEraSeq.FindByEraAndStashAccount(payload.Syncable.Era, rawValidatorEraSeq.StashAccount)
		if err != nil {
			if err == store.ErrNotFound {
				newValidatorEraSeqs = append(newValidatorEraSeqs, rawValidatorEraSeq)
				continue
			} else {
				return err
			}
		}

		validatorEraSeq.Update(rawValidatorEraSeq)
		updatedValidatorEraSeqs = append(updatedValidatorEraSeqs, *validatorEraSeq)
	}

	payload.NewValidatorEraSequences = newValidatorEraSeqs
	payload.UpdatedValidatorEraSequences = updatedValidatorEraSeqs

	return nil
}

// NewEventSeqCreatorTask creates block sequences
func NewEventSeqCreatorTask(db *store.Store) *eventSeqCreatorTask {
	return &eventSeqCreatorTask{
		db: db,
	}
}

type eventSeqCreatorTask struct {
	db *store.Store
}

func (t *eventSeqCreatorTask) GetName() string {
	return EventSeqCreatorTaskName
}

func (t *eventSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	mappedEventSeqs, err := EventToSequence(payload.Syncable, payload.RawEvents)
	if err != nil {
		return err
	}

	var newEventSeqs []model.EventSeq
	var updatedEventSeqs []model.EventSeq
	for _, rawEventSeq := range mappedEventSeqs {
		validatorSeq, err := t.db.EventSeq.FindByHeightAndIndex(payload.CurrentHeight, rawEventSeq.Index)
		if err != nil {
			if err == store.ErrNotFound {
				newEventSeqs = append(newEventSeqs, rawEventSeq)
				continue
			} else {
				return err
			}
		}

		validatorSeq.Update(rawEventSeq)
		updatedEventSeqs = append(updatedEventSeqs, *validatorSeq)
	}

	payload.NewEventSequences = newEventSeqs
	payload.UpdatedEventSequences = updatedEventSeqs

	return nil
}