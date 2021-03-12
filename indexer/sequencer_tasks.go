package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
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

	eventMethodReward     = "Reward"
	txMethodPayoutStakers = "payoutStakers"
	txMethodBatch         = "batch"
	txMethodBatchAll      = "batchAll"
	txMethodProxy         = "proxy"
	txMethodSudo          = "sudo"
	sectionUtility        = "utility"
	sectionStaking        = "staking"
	sectionProxy          = "proxy"

	accountKey = "AccountId"
	balanceKey = "Balance"
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

	mappedTxSeqs, err := ToTransactionSequence(payload.Syncable, payload.RawBlock.GetExtrinsics())
	if err != nil {
		return err
	}

	payload.TransactionSequences = mappedTxSeqs

	return nil
}

// NewRewardEraSeqCreatorTask creates rewards
func NewRewardEraSeqCreatorTask(cfg *config.Config, rewardsDb store.Rewards, syncablesDb store.Syncables, validatorDb store.ValidatorEraSeq) *rewardEraSeqCreatorTask {
	return &rewardEraSeqCreatorTask{
		cfg: cfg,

		rewardsDb:   rewardsDb,
		syncablesDb: syncablesDb,
		validatorDb: validatorDb,
	}
}

type rewardEraSeqCreatorTask struct {
	cfg *config.Config

	rewardsDb   store.Rewards
	syncablesDb store.Syncables
	validatorDb store.ValidatorEraSeq
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

	// get unclaimed rewards for validators in current era
	for stash, validator := range payload.ParsedValidators {
		data := validator.parsedRewards

		if data.Commission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &currentEraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Commission,
				Kind:                  model.RewardCommission,
				Claimed:               data.IsClaimed,
			})
		}

		if data.Reward != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &currentEraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Reward,
				Kind:                  model.RewardReward,
				Claimed:               data.IsClaimed,
			})
		}

		if data.RewardAndCommission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &currentEraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.RewardAndCommission,
				Kind:                  model.RewardCommissionAndReward,
				Claimed:               data.IsClaimed,
			})
		}

		for _, n := range data.StakerRewards {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &currentEraSeq,
				StashAccount:          n.Stash,
				ValidatorStashAccount: stash,
				Amount:                n.Amount,
				Kind:                  model.RewardReward,
				Claimed:               data.IsClaimed,
			})
		}
	}

	// get claimed rewards from staking.payoutStakers txs
	for _, tx := range payload.RawBlock.GetExtrinsics() {
		var claims []RewardsClaim
		var err error

		if !tx.GetIsSuccess() {
			continue
		}

		if tx.GetSection() == sectionStaking && tx.GetMethod() == txMethodPayoutStakers {
			claim, err := getRewardsClaimFromPayoutStakersTx(tx.GetArgs())
			if err != nil {
				return err
			}
			claim.TxHash = tx.GetHash()
			claims = append(claims, claim)
		} else if (tx.GetSection() == sectionUtility && (tx.GetMethod() == txMethodBatch || tx.GetMethod() == txMethodBatchAll)) ||
			(tx.GetSection() == sectionProxy && tx.GetMethod() == txMethodProxy) {
			for _, callArg := range tx.GetCallArgs() {
				if callArg.GetSection() == sectionStaking && callArg.GetMethod() == txMethodPayoutStakers {
					claim, err := getRewardsClaimFromPayoutStakersTx(callArg.GetValue())
					if err != nil {
						return err
					}
					claim.TxHash = tx.GetHash()
					claims = append(claims, claim)
				}
			}
		}

		rewards, rewardclaims, err := t.getRewardsFromEvents(tx.GetExtrinsicIndex(), claims, payload.RawEvents)
		if err != nil {
			return err
		}

		payload.RewardsClaimed = append(payload.RewardsClaimed, rewardclaims...)
		payload.RewardEraSequences = append(payload.RewardEraSequences, rewards...)
	}

	return nil
}

func (t *rewardEraSeqCreatorTask) getEraSeq(era int64, currentSyncable *model.Syncable) (model.EraSequence, error) {
	var firstHeightInEra int64
	lastSyncableInPrevEra, err := t.syncablesDb.FindLastInEra(era - 1)
	if err != nil {
		if err != store.ErrNotFound {
			return model.EraSequence{}, err
		}
		firstHeightInEra = t.cfg.FirstBlockHeight
	} else {
		firstHeightInEra = lastSyncableInPrevEra.Height + 1
	}

	lastSyncableInEra := currentSyncable
	if currentSyncable == nil || currentSyncable.Era != era {
		lastSyncableInEra, err = t.syncablesDb.FindLastInEra(era)
		if err != nil {
			return model.EraSequence{}, err
		}
	}

	return model.EraSequence{
		Era:         era,
		StartHeight: firstHeightInEra,
		EndHeight:   lastSyncableInEra.Height + 1,
		Time:        lastSyncableInEra.Time,
	}, nil
}

// getRewardsFromEvents rreturns claims if rewards exist already in db, and returns new reward seqs if reward seqs don't exist in db
func (t *rewardEraSeqCreatorTask) getRewardsFromEvents(txIdx int64, claims []RewardsClaim, events []*eventpb.Event) ([]model.RewardEraSeq, []RewardsClaim, error) {
	var rewards []model.RewardEraSeq
	var rewardClaims []RewardsClaim

	if len(events) == 0 && len(claims) != 0 {
		return rewards, rewardClaims, fmt.Errorf("[getRewardsFromEvents] expected events to not be empty")
	}

	var idx int
	var nextVald string
	for i, claim := range claims {
		if i < len(claims)-1 {
			nextVald = claims[i+1].ValidatorStash
		} else {
			nextVald = ""
		}

		eraSeq, err := t.getEraSeq(claim.Era, nil)
		if err != nil {
			return rewards, rewardClaims, err
		}

		extractedRewards, ranged, err := t.extractRewardsForClaimFromEvents(claim, eraSeq, txIdx, nextVald, events[idx:])
		if err != nil {
			return rewards, rewardClaims, err
		}
		idx += ranged

		count, err := t.rewardsDb.GetCount(claim.ValidatorStash, claim.Era)
		if err != nil {
			return rewards, rewardClaims, err
		}

		if count == 0 {
			rewards = append(rewards, extractedRewards...)
			continue
		}

		// if already exists in db, then claim was added to RewardsClaimed, so don't add parsedRewards
		// instead check that count in db matches parsedRewards. Count may be +1 more since unclaimed validator rewards
		// can be split into commission and reward
		if int(count) != len(extractedRewards) && int(count) != len(extractedRewards)+1 {
			return rewards, rewardClaims, fmt.Errorf("Expected unclaimed rewards to match number of rewards from claim %v; got: %v want: %v (~-1)", claim, len(extractedRewards), count)
		} else {
			rewardClaims = append(rewardClaims, claim)
		}
	}

	return rewards, rewardClaims, nil
}

func (t *rewardEraSeqCreatorTask) extractRewardsForClaimFromEvents(claim RewardsClaim, eraSeq model.EraSequence, txIdx int64, nextVald string, events []*eventpb.Event) ([]model.RewardEraSeq, int, error) {
	var rewards []model.RewardEraSeq
	var foundCurrVald bool

	for i, event := range events {
		if event.GetExtrinsicIndex() != txIdx || event.GetMethod() != eventMethodReward || event.GetSection() != sectionStaking {
			continue
		}

		stash, amount, err := t.getStashAndAmountFromData(event)
		if err != nil {
			return rewards, i, err
		}

		if foundCurrVald && stash == nextVald {
			return rewards, i, nil
		}

		if stash == claim.ValidatorStash {
			foundCurrVald = true
		}

		reward := model.RewardEraSeq{
			Kind:                  model.RewardReward,
			EraSequence:           &eraSeq,
			StashAccount:          stash,
			Amount:                amount.String(),
			ValidatorStashAccount: claim.ValidatorStash,
			Claimed:               true,
			TxHash:                claim.TxHash,
		}

		if stash == claim.ValidatorStash {
			reward.Kind = model.RewardCommissionAndReward
		}

		rewards = append(rewards, reward)
	}

	if nextVald != "" {
		return rewards, 0, fmt.Errorf("Could not extract rewards from events")
	}

	return rewards, 0, nil
}

func (t *rewardEraSeqCreatorTask) getStashAndAmountFromData(event *eventpb.Event) (stash string, amount types.Quantity, err error) {
	for _, d := range event.GetData() {
		if d.GetName() == accountKey {
			stash = d.Value
		} else if d.GetName() == balanceKey {
			amount, err = types.NewQuantityFromString(d.Value)
		}
	}
	if stash == "" {
		err = errUnexpectedEventDataFormat
	}
	return
}

func getRewardsClaimFromPayoutStakersTx(args string) (RewardsClaim, error) {
	var data []string

	err := json.Unmarshal([]byte(args), &data)
	if err == nil {
		return getStashAndEraFromPayoutArgs(data)
	}

	parts := strings.Split(args, ",")
	if len(parts) != 2 {
		return RewardsClaim{}, errUnexpectedEventDataFormat
	}

	return getStashAndEraFromPayoutArgs(parts)
}

func getStashAndEraFromPayoutArgs(data []string) (RewardsClaim, error) {
	if len(data) < 2 {
		return RewardsClaim{}, errUnexpectedTxDataFormat
	}

	era, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return RewardsClaim{}, err
	}

	return RewardsClaim{
		Era:            era,
		ValidatorStash: data[0],
	}, nil
}
