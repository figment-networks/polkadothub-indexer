package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
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

	errCannotCalculateRewards = errors.New("cannot calculate rewards")
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
	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	eraSeq, err := t.getEraSeq(payload.HeightMeta.ActiveEra, payload.Syncable)
	if err != nil {
		return err
	}

	// get unclaimed rewards for validators in current era
	for stash, validator := range payload.ParsedValidators {
		data := validator.parsedRewards

		if data.Commission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Commission,
				Kind:                  model.RewardCommission,
				Claimed:               data.IsClaimed,
			})
		}

		if data.Reward != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.Reward,
				Kind:                  model.RewardReward,
				Claimed:               data.IsClaimed,
			})
		}

		if data.RewardAndCommission != "" {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &eraSeq,
				StashAccount:          stash,
				ValidatorStashAccount: stash,
				Amount:                data.RewardAndCommission,
				Kind:                  model.RewardCommissionAndReward,
				Claimed:               data.IsClaimed,
			})
		}

		for _, n := range data.StakerRewards {
			payload.RewardEraSequences = append(payload.RewardEraSequences, model.RewardEraSeq{
				EraSequence:           &eraSeq,
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

		if len(claims) == 0 {
			continue
		}

		legitimateClaims, rewardArgs, err := t.getLegitimateClaimsAndRewardArgs(claims, payload.RawEvents, tx.GetExtrinsicIndex())
		if err != nil {
			if errors.Is(err, errCannotCalculateRewards) {
				// impossible to extract rewards for this claim, so report error and keep going (should be calulcated successfully via backfill)
				err = fmt.Errorf("%w: height: %d", err, payload.CurrentHeight)
				logger.Error(err, logger.Field("height", payload.CurrentHeight), logger.Field("task", t.GetName()), logger.Field("stage", pipeline.StageSequencer))
				continue
			}
			return err
		}

		rewards, rewardclaims, err := t.extractRewards(legitimateClaims, rewardArgs)
		if err != nil {
			if errors.Is(err, errCannotCalculateRewards) {
				// don't stop pipeline on error, report error to investigate later and keep going
				err = fmt.Errorf("%w: height: %d", err, payload.CurrentHeight)
				logger.Error(err, logger.Field("height", payload.CurrentHeight), logger.Field("task", t.GetName()), logger.Field("stage", pipeline.StageSequencer))
				continue
			}
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

// getLegitimateClaimsAndRewardArgs filters claims out claims that are skipped by polkadot
// and checks that all claims have accompanying reward events.
// Makes the following assumptions:
// a) if validator has already claimed rewards for era, then expect batchInterrupted error 'Rewards for this era have already been claimed for this validator' (see 0x1c9708278cad4caf0fa8b95510ceba627f232a540de3dfea58b09aae78b1e44b)
// b) if claim contains invalid era, then expect batchInterrupted error 'Invalid era to reward' (see 0xa4f468cda9e5dd7b290da35a786e53b2b704bc66196c4336aeba91a6a8cc0b6d)
// c) if claim contains invalid validator (not an era validator, or not enough reward points for era), then polkadot will skip claim without erroring
func (t *rewardEraSeqCreatorTask) getLegitimateClaimsAndRewardArgs(claims []RewardsClaim, events []*eventpb.Event, txIdx int64) ([]RewardsClaim, []rewardEventArgs, error) {
	var legitimate []RewardsClaim
	var args []rewardEventArgs

	// filter out claims that are ignored/skipped by polkadot
	filteredClaims := []RewardsClaim{}
	for _, c := range claims {
		validator, err := t.validatorDb.FindByEraAndStashAccount(c.Era, c.ValidatorStash)
		if err == store.ErrNotFound {
			// if claim contains invalid validator, then polkadot will skip claim (see 0x1c9708278cad4caf0fa8b95510ceba627f232a540de3dfea58b09aae78b1e44b, idx 5 is not a validator for claim era)
			continue
		} else if err != nil {
			return legitimate, args, err
		}

		// if validator doesn't have reward points for era, then claim will be skipped (see 12vCBNjJrXMpJPBh1Mr366CFp2nLCp1AFJAp6LryZFchFPKn, idx 0)
		if validator.RewardPoints == 0 {
			continue
		}
		filteredClaims = append(filteredClaims, c)
	}

	var currentValIdx int
	var currentClaim RewardsClaim
	if len(filteredClaims) > 0 {
		currentClaim = filteredClaims[currentValIdx]
	}

	var batchInterrupted bool
	for _, ev := range events {
		if ev.GetExtrinsicIndex() != txIdx {
			continue
		}

		if ev.GetMethod() == "BatchInterrupted" {
			batchInterrupted = true
			// check idx of error
			claimIdx, err := t.getErrorIndexFromBatchInterruptedData(ev)
			if err != nil {
				return legitimate, args, err
			}
			if claimIdx > len(claims) || claims[claimIdx].ValidatorStash != currentClaim.ValidatorStash {
				return legitimate, args, fmt.Errorf("BatchInterrupted event data contains unexpected claim index: %w", errCannotCalculateRewards)
			}
			break
		}

		if ev.GetMethod() != eventMethodReward || ev.GetSection() != sectionStaking {
			continue
		}

		arg, err := t.getRewardEventArgsFromData(ev)
		if err != nil {
			return legitimate, args, err
		}
		args = append(args, arg)

		if arg.stash != currentClaim.ValidatorStash {
			continue
		}

		legitimate = append(legitimate, currentClaim)

		currentValIdx++
		if currentValIdx >= len(filteredClaims) {
			currentClaim = RewardsClaim{}
			continue
		}
		currentClaim = filteredClaims[currentValIdx]
	}

	if currentClaim.ValidatorStash != "" && !batchInterrupted {
		return legitimate, args, fmt.Errorf("Expected to find reward events for claim %v %v: %w", currentClaim.ValidatorStash, currentClaim.Era, errCannotCalculateRewards)
	}

	if len(filteredClaims) == 0 && len(args) > 0 {
		return legitimate, args, fmt.Errorf("Got reward event but no claim: %w", errCannotCalculateRewards)
	}

	return legitimate, args, nil
}

// extractRewards rreturns claims if rewards exist already in db, and returns new reward seqs if reward seqs don't exist in db
func (t *rewardEraSeqCreatorTask) extractRewards(claims []RewardsClaim, rewardArgs []rewardEventArgs) ([]model.RewardEraSeq, []RewardsClaim, error) {
	var rewards []model.RewardEraSeq
	var rewardClaims []RewardsClaim

	if len(rewardArgs) == 0 && len(claims) != 0 {
		return rewards, rewardClaims, fmt.Errorf("expected rewardArgs to not be empty: %w", errCannotCalculateRewards)
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

		extractedRewards, ranged, err := t.extractRewardsForClaimFromRewardEventArgs(claim, eraSeq, nextVald, rewardArgs[idx:])
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
			// todo only mark extractedRewards rewards as claimed, leave calculated rewards from db as unclaimed
			return rewards, rewardClaims, fmt.Errorf("Expected unclaimed rewards to match number of rewards from claim %v; got: %v want: %v (~-1): %w", claim, len(extractedRewards), count, errCannotCalculateRewards)
		} else {
			rewardClaims = append(rewardClaims, claim)
		}
	}

	return rewards, rewardClaims, nil
}

func (t *rewardEraSeqCreatorTask) extractRewardsForClaimFromRewardEventArgs(claim RewardsClaim, eraSeq model.EraSequence, nextVald string, rewardArgs []rewardEventArgs) ([]model.RewardEraSeq, int, error) {
	var rewards []model.RewardEraSeq
	var foundCurrVald bool

	for i, arg := range rewardArgs {
		if foundCurrVald && arg.stash == nextVald {
			return rewards, i, nil
		}

		if arg.stash == claim.ValidatorStash {
			foundCurrVald = true
		}

		reward := model.RewardEraSeq{
			Kind:                  model.RewardReward,
			EraSequence:           &eraSeq,
			StashAccount:          arg.stash,
			Amount:                arg.amount,
			ValidatorStashAccount: claim.ValidatorStash,
			Claimed:               true,
			TxHash:                claim.TxHash,
		}

		if arg.stash == claim.ValidatorStash {
			reward.Kind = model.RewardCommissionAndReward
		}

		rewards = append(rewards, reward)
	}

	if nextVald != "" {
		return rewards, 0, fmt.Errorf("Could not extract rewards from events")
	}

	return rewards, 0, nil
}

type rewardEventArgs struct {
	stash  string
	amount string
}

func (t *rewardEraSeqCreatorTask) getRewardEventArgsFromData(event *eventpb.Event) (args rewardEventArgs, err error) {
	for _, d := range event.GetData() {
		switch d.GetName() {
		case accountKey:
			args.stash = d.Value
		case balanceKey:
			args.amount = d.Value
		}
	}
	if args.stash == "" {
		err = errUnexpectedEventDataFormat
	}
	return
}

func (t *rewardEraSeqCreatorTask) getErrorIndexFromBatchInterruptedData(event *eventpb.Event) (index int, err error) {
	var foundIndex bool
	var isDispatchErr bool
	for _, d := range event.GetData() {
		switch d.GetName() {
		case "u32":
			index, err = strconv.Atoi(d.Value)
			if err != nil {
				return -1, err
			}
			foundIndex = true
		case "DispatchError":
			isDispatchErr = true
		}
	}
	if !isDispatchErr || !foundIndex {
		return -1, fmt.Errorf("unexpected format for BatchInterrupted event data")
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
