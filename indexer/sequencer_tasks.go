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
	"github.com/figment-networks/polkadothub-indexer/types"
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

	if !payload.Syncable.LastInEra {
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

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

	eraSeq := &model.EraSequence{
		Era:         payload.Syncable.Era,
		StartHeight: firstHeightInEra,
		EndHeight:   payload.Syncable.Height,
		Time:        payload.Syncable.Time,
	}

	for stash, data := range payload.ParsedValidators {
		if data.UnclaimedCommission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.UnclaimedCommission,
				Kind:                  model.RewardCommission,
				Claimed:               false,
			})
		}

		if data.UnclaimedReward != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.UnclaimedReward,
				Kind:                  model.RewardReward,
				Claimed:               false,
			})
		}

		for _, n := range data.UnclaimedStakerRewards {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           eraSeq,
				StashAccount:          n.Stash,
				ValidatorStashAccount: stash,
				Amount:                n.Amount,
				Kind:                  model.RewardReward,
				Claimed:               false,
			})
		}
	}

	return nil
}

// NewClaimedRewardEraSeqCreatorTask creates rewards
func NewClaimedRewardEraSeqCreatorTask(cfg *config.Config, rewardsDb store.Rewards, syncablesDb store.Syncables, validatorDb store.ValidatorEraSeq) *claimedRewardEraSeqCreatorTask {
	return &claimedRewardEraSeqCreatorTask{cfg, rewardsDb, syncablesDb, validatorDb}
}

type claimedRewardEraSeqCreatorTask struct {
	cfg         *config.Config
	rewardsDb   store.Rewards
	syncablesDb store.Syncables
	validatorDb store.ValidatorEraSeq
}

func (t *claimedRewardEraSeqCreatorTask) GetName() string {
	return ClaimedRewardEraSeqCreatorTaskName
}

func (t *claimedRewardEraSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)
	for _, tx := range payload.TransactionSequences {
		if !tx.IsPayoutStakers() {
			continue
		}

		validatorStash, era, err := tx.GetStashAndEraFromPayoutArgs()
		if err != nil {
			return err
		}

		//check if already exists in db, if yes mark all claimed
		count, err := t.rewardsDb.GetCount(validatorStash, era)
		if err != nil {
			return err
		}

		if count == 0 {
			// these are historical rewards whose unclaimed reward data is not in the database
			rewards, err := t.getRewardsFromEvents(validatorStash, era, payload.EventSequences)
			if err != nil {
				return err
			}
			payload.RewardEraSequences = append(payload.RewardEraSequences, rewards...)
			continue
		}

		payload.RewardsClaimed = append(payload.RewardsClaimed, RewardsClaim{
			Era:            era,
			ValidatorStash: validatorStash,
		})
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

	return nil
}

func (t *claimedRewardEraSeqCreatorTask) getEraSeq(era int64) (*model.EraSequence, error) {
	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.syncablesDb.FindLastInEra(era - 1)
	if err != nil {
		if err == store.ErrNotFound {
			firstHeightInEra = t.cfg.FirstBlockHeight
		} else {
			return nil, err
		}
	} else {
		firstHeightInEra = lastSyncableInPrevEra.Height + 1
	}

	lastSyncableInEra, err := t.syncablesDb.FindLastInEra(era)
	if err != nil {
		return nil, err
	}

	return &model.EraSequence{
		Era:         era,
		StartHeight: firstHeightInEra,
		EndHeight:   lastSyncableInEra.Height + 1,
		Time:        lastSyncableInEra.Time,
	}, nil
}

func (t *claimedRewardEraSeqCreatorTask) getRewardsFromEvents(validatorStash string, era int64, events []model.EventSeq) ([]model.RewardEraSeq, error) {
	var rewardSeqs []model.RewardEraSeq

	eraSeq, err := t.getEraSeq(era)
	if err != nil {
		return nil, err
	}

	var totalStakerRewards types.Quantity
	var validatorRewardAndCommission types.Quantity

	for _, event := range events {
		if !event.IsReward() {
			continue
		}

		stash, amount, err := event.GetStashAndAmountFromData()
		if err != nil {
			return nil, err
		}

		if stash == validatorStash {
			validatorRewardAndCommission = amount
			continue
		}

		rewardSeqs = append(rewardSeqs, model.RewardEraSeq{
			EraSequence:           eraSeq,
			StashAccount:          stash,
			ValidatorStashAccount: validatorStash,
			Amount:                amount.String(),
			Kind:                  model.RewardReward,
			Claimed:               true,
		})

		totalStakerRewards.Add(amount)
	}

	if validatorRewardAndCommission.Cmp(&zero) <= 0 {
		return rewardSeqs, nil
	}

	validator, err := t.validatorDb.FindByEraAndStashAccount(era, validatorStash)
	if err != nil {
		return nil, err
	}

	if validator.Commission == hundredpermill || validator.OwnStake.Cmp(&zero) <= 0 {
		return append(rewardSeqs, model.RewardEraSeq{
			EraSequence:           eraSeq,
			StashAccount:          validatorStash,
			ValidatorStashAccount: validatorStash,
			Amount:                validatorRewardAndCommission.String(),
			Kind:                  model.RewardCommission,
			Claimed:               true,
		}), nil
	}

	if validator.Commission == 0 {
		return append(rewardSeqs, model.RewardEraSeq{
			EraSequence:           eraSeq,
			StashAccount:          validatorStash,
			ValidatorStashAccount: validatorStash,
			Amount:                validatorRewardAndCommission.String(),
			Kind:                  model.RewardReward,
			Claimed:               true,
		}), nil
	}

	return append(rewardSeqs, model.RewardEraSeq{
		EraSequence:           eraSeq,
		StashAccount:          validatorStash,
		ValidatorStashAccount: validatorStash,
		Amount:                validatorRewardAndCommission.String(),
		Kind:                  model.RewardCommissionAndReward,
		Claimed:               true,
	}), nil
}
