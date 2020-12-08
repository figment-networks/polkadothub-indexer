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
	ValidatorSeqCreatorTaskName        = "ValidatorSeqCreator"
	ValidatorSessionSeqCreatorTaskName = "ValidatorSessionSeqCreator"
	ValidatorEraSeqCreatorTaskName     = "ValidatorEraSeqCreator"
	EventSeqCreatorTaskName            = "EventSeqCreator"
	AccountEraSeqCreatorTaskName       = "AccountEraSeqCreator"
	TransactionSeqCreatorTaskName      = "TransactionSeqCreator"
	RewardEraSeqCreatorTaskName        = "RewardEraSeqCreator"
	ClaimedRewardEraSeqCreatorTaskName = "ClaimedRewardEraSeqCreator"
)

var (
	_ pipeline.Task = (*blockSeqCreatorTask)(nil)
	_ pipeline.Task = (*validatorSessionSeqCreatorTask)(nil)
)

const (
	hundredpermill = 1000000000
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

// NewValidatorSeqCreatorTask creates validator sequences
func NewValidatorSeqCreatorTask(validatorSeqDb store.ValidatorSeq) *validatorSeqCreatorTask {
	return &validatorSeqCreatorTask{
		validatorSeqDb: validatorSeqDb,
	}
}

type validatorSeqCreatorTask struct {
	validatorSeqDb store.ValidatorSeq
}

func (t *validatorSeqCreatorTask) GetName() string {
	return ValidatorSeqCreatorTaskName
}

func (t *validatorSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	mappedValidatorSeqs, err := ToValidatorSequence(payload.Syncable, payload.RawValidators)
	if err != nil {
		return err
	}

	payload.ValidatorSequences = mappedValidatorSeqs
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

	payload.ValidatorSessionSequences = mappedValidatorSessionSeqs
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

	payload.ValidatorEraSequences = mappedValidatorEraSeqs
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

	payload.EventSequences = mappedEventSeqs
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

	for _, stakingValidator := range payload.RawStaking.GetValidators() {
		mappedAccountEraSeqs, err := ToAccountEraSequence(payload.Syncable, firstHeightInEra, stakingValidator)
		if err != nil {
			return err
		}

		payload.AccountEraSequences = append(payload.AccountEraSequences, mappedAccountEraSeqs...)
	}

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

	payload.TransactionSequences = mappedTxSeqs

	return nil
}

// NewRewardEraSeqCreatorTask creates rewards
func NewRewardEraSeqCreatorTask(cfg *config.Config, syncablesDb store.Syncables) *rewardEraSeqCreatorTask {
	return &rewardEraSeqCreatorTask{cfg, syncablesDb}
}

type rewardEraSeqCreatorTask struct {
	cfg         *config.Config
	syncablesDb store.Syncables
}

func (t *rewardEraSeqCreatorTask) GetName() string {
	return RewardEraSeqCreatorTaskName
}

func (t *rewardEraSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

	currentEraSeq, err := t.getEraSeq(payload.Syncable.Era, payload.Syncable)
	if err != nil {
		return err
	}

	var eraSeq *model.EraSequence
	for stash, validator := range payload.ParsedValidators {
		data := validator.parsedRewards

		eraSeq = currentEraSeq
		if data.Era != payload.Syncable.Era {
			eraSeq, err = t.getEraSeq(data.Era, payload.Syncable)
			if err != nil {
				return err
			}
		}

		if data.Commission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Commission,
				Kind:                  model.RewardCommission,
				Claimed:               data.IsClaimed,
			})
		}

		if data.Reward != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Reward,
				Kind:                  model.RewardReward,
				Claimed:               data.IsClaimed,
			})
		}

		if data.RewardAndCommission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.RewardAndCommission,
				Kind:                  model.RewardCommissionAndReward,
				Claimed:               data.IsClaimed,
			})
		}

		for _, n := range data.StakerRewards {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          n.Stash,
				ValidatorStashAccount: stash,
				Amount:                n.Amount,
				Kind:                  model.RewardReward,
				Claimed:               data.IsClaimed,
			})
		}
	}

	return nil
}

func (t *rewardEraSeqCreatorTask) getEraSeq(era int64, currentSyncable *model.Syncable) (*model.EraSequence, error) {
	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.syncablesDb.FindLastInEra(era - 1)
	if err != nil {
		if err != store.ErrNotFound {
			return nil, err
		}
		firstHeightInEra = t.cfg.FirstBlockHeight
	} else {
		firstHeightInEra = lastSyncableInPrevEra.Height + 1
	}

	lastSyncableInEra := currentSyncable
	if currentSyncable.Era != era {
		lastSyncableInEra, err = t.syncablesDb.FindLastInEra(era)
		if err != nil {
			return nil, err
		}
	}

	return &model.EraSequence{
		Era:         era,
		StartHeight: firstHeightInEra,
		EndHeight:   lastSyncableInEra.Height + 1,
		Time:        lastSyncableInEra.Time,
	}, nil
}
