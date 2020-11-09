package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	BlockSeqCreatorTaskName            = "BlockSeqCreator"
	ValidatorSessionSeqCreatorTaskName = "ValidatorSessionSeqCreator"
	ValidatorEraSeqCreatorTaskName     = "ValidatorEraSeqCreator"
	EventSeqCreatorTaskName            = "EventSeqCreator"
	AccountEraSeqCreatorTaskName       = "AccountEraSeqCreator"
	TransactionSeqCreatorTaskName      = "TransactionSeqCreator"
)

var (
	_ pipeline.Task = (*blockSeqCreatorTask)(nil)
	_ pipeline.Task = (*validatorSessionSeqCreatorTask)(nil)
)

// NewBlockSeqCreatorTask creates block sequences
func NewBlockSeqCreatorTask(blockSeqDb store.BlockSeq) *blockSeqCreatorTask {
	return &blockSeqCreatorTask{
		blockSeqDb: blockSeqDb,
	}
}

type blockSeqCreatorTask struct {
	blockSeqDb store.BlockSeq
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

	blockSeq, err := t.blockSeqDb.FindSeqByHeight(payload.CurrentHeight)
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
func NewValidatorSessionSeqCreatorTask(cfg *config.Config, syncablesDb store.Syncables, validatorSessionSeqDb store.ValidatorSessionSeq) *validatorSessionSeqCreatorTask {
	return &validatorSessionSeqCreatorTask{
		cfg:                   cfg,
		syncablesDb:           syncablesDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
	}
}

type validatorSessionSeqCreatorTask struct {
	cfg                   *config.Config
	syncablesDb           store.Syncables
	validatorSessionSeqDb store.ValidatorSessionSeq
}

func (t *validatorSessionSeqCreatorTask) GetName() string {
	return ValidatorSessionSeqCreatorTaskName
}

func (t *validatorSessionSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in session [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	var firstHeightInSession int64
	lastSyncableInPrevSession, err := t.syncablesDb.FindLastInSession(payload.Syncable.Session - 1)
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
		validatorSessionSeq, err := t.validatorSessionSeqDb.FindBySessionAndStashAccount(payload.Syncable.Session, rawValidatorSessionSeq.StashAccount)
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
func NewValidatorEraSeqCreatorTask(cfg *config.Config, syncablesDb store.Syncables, validatorEraSeqDb store.ValidatorEraSeq) *validatorEraSeqCreatorTask {
	return &validatorEraSeqCreatorTask{
		cfg:               cfg,
		syncablesDb:       syncablesDb,
		validatorEraSeqDb: validatorEraSeqDb,
	}
}

type validatorEraSeqCreatorTask struct {
	cfg               *config.Config
	syncablesDb       store.Syncables
	validatorEraSeqDb store.ValidatorEraSeq
}

func (t *validatorEraSeqCreatorTask) GetName() string {
	return ValidatorEraSeqCreatorTaskName
}

func (t *validatorEraSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.syncablesDb.FindLastInEra(payload.Syncable.Era - 1)
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
		validatorEraSeq, err := t.validatorEraSeqDb.FindByEraAndStashAccount(payload.Syncable.Era, rawValidatorEraSeq.StashAccount)
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
func NewEventSeqCreatorTask(eventSeqDb store.EventSeq) *eventSeqCreatorTask {
	return &eventSeqCreatorTask{
		eventSeqDb: eventSeqDb,
	}
}

type eventSeqCreatorTask struct {
	eventSeqDb store.EventSeq
}

func (t *eventSeqCreatorTask) GetName() string {
	return EventSeqCreatorTaskName
}

func (t *eventSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	mappedEventSeqs, err := ToEventSequence(payload.Syncable, payload.RawEvents)
	if err != nil {
		return err
	}

	txIndexes := make([]int64, len(mappedEventSeqs))
	for i, seq := range mappedEventSeqs {
		txIndexes[i] = seq.Index
	}

	seqLookup, err := t.eventSeqDb.FindAllByHeightAndIndex(payload.CurrentHeight, txIndexes)
	if err != nil {
		return err
	}

	var newEventSeqs []model.EventSeq
	var updatedEventSeqs []model.EventSeq
	for _, rawEventSeq := range mappedEventSeqs {
		seq, exists := seqLookup[rawEventSeq.Index]
		if !exists {
			newEventSeqs = append(newEventSeqs, rawEventSeq)
			continue
		}

		seq.Update(rawEventSeq)
		updatedEventSeqs = append(updatedEventSeqs, *seq)
	}

	payload.NewEventSequences = newEventSeqs
	payload.UpdatedEventSequences = updatedEventSeqs

	return nil
}

// NewAccountEraSeqCreatorTask creates account era sequences
func NewAccountEraSeqCreatorTask(cfg *config.Config, accountEraSeqDb store.AccountEraSeq, syncablesDb store.Syncables) *accountEraSeqCreatorTask {
	return &accountEraSeqCreatorTask{
		cfg:             cfg,
		accountEraSeqDb: accountEraSeqDb,
		syncablesDb:     syncablesDb,
	}
}

type accountEraSeqCreatorTask struct {
	cfg             *config.Config
	accountEraSeqDb store.AccountEraSeq
	syncablesDb     store.Syncables
}

func (t *accountEraSeqCreatorTask) GetName() string {
	return AccountEraSeqCreatorTaskName
}

func (t *accountEraSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.syncablesDb.FindLastInEra(payload.Syncable.Era - 1)
	if err != nil {
		if err == store.ErrNotFound {
			firstHeightInEra = t.cfg.FirstBlockHeight
		} else {
			return err
		}
	} else {
		firstHeightInEra = lastSyncableInPrevEra.Height + 1
	}

	var newAccountEraSeqs []model.AccountEraSeq
	var updatedAccountEraSeqs []model.AccountEraSeq
	for _, stakingValidator := range payload.RawStaking.GetValidators() {
		mappedAccountEraSeqs, err := ToAccountEraSequence(payload.Syncable, firstHeightInEra, stakingValidator)
		if err != nil {
			return err
		}

		for _, mappedAccountEraSeq := range mappedAccountEraSeqs {
			validatorEraSeq, err := t.accountEraSeqDb.FindByEraAndStashAccounts(payload.Syncable.Era, mappedAccountEraSeq.StashAccount, mappedAccountEraSeq.StashAccount)
			if err != nil {
				if err == store.ErrNotFound {
					newAccountEraSeqs = append(newAccountEraSeqs, mappedAccountEraSeq)
					continue
				} else {
					return err
				}
			}

			validatorEraSeq.Update(mappedAccountEraSeq)
			updatedAccountEraSeqs = append(updatedAccountEraSeqs, *validatorEraSeq)
		}
	}
	payload.NewAccountEraSequences = newAccountEraSeqs
	payload.UpdatedAccountEraSequences = updatedAccountEraSeqs

	return nil
}

// NewTransactionSeqCreatorTask creates block sequences
func NewTransactionSeqCreatorTask(transactionSeqDb store.TransactionSeq) *transactionSeqCreatorTask {
	return &transactionSeqCreatorTask{
		transactionSeqDb: transactionSeqDb,
	}
}

type transactionSeqCreatorTask struct {
	transactionSeqDb store.TransactionSeq
}

func (t *transactionSeqCreatorTask) GetName() string {
	return TransactionSeqCreatorTaskName
}

func (t *transactionSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	mappedTxSeqs, err := ToTransactionSequence(payload.Syncable, payload.RawTransactions)
	if err != nil {
		return err
	}

	txIndexes := make([]int64, len(mappedTxSeqs))
	for i, seq := range mappedTxSeqs {
		txIndexes[i] = seq.Index
	}

	seqLookup, err := t.transactionSeqDb.FindAllByHeightAndIndex(payload.CurrentHeight, txIndexes)
	if err != nil {
		return err
	}

	var newTxSeqs []model.TransactionSeq
	var updatedTxSeqs []model.TransactionSeq
	for _, rawSeq := range mappedTxSeqs {
		seq, exists := seqLookup[rawSeq.Index]
		if !exists {
			newTxSeqs = append(newTxSeqs, rawSeq)
			continue
		}

		seq.Update(rawSeq)
		updatedTxSeqs = append(updatedTxSeqs, *seq)
	}

	payload.NewTransactionSequences = newTxSeqs
	payload.UpdatedTransactionSequences = updatedTxSeqs

	return nil
}
