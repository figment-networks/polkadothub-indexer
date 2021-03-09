package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
)

// (3001567,batch,utility)
// (3001801,batchAll,utility)
//  (3046637,proxy,proxy)

const (
	BlockParserTaskName       = "BlockParser"
	ValidatorsParserTaskName  = "ValidatorsParser"
	TransactionParserTaskName = "TransactionParser"

	eventMethodReward     = "Reward"
	txMethodPayoutStakers = "payoutStakers"
	txMethodBatch         = "batch"
	txMethodBatchAll      = "batchAll"
	sectionUtility        = "utility"
	sectionStaking        = "staking"

	accountKey = "AccountId"
	balanceKey = "Balance"
)

var (
	_ pipeline.Task = (*blockParserTask)(nil)
	_ pipeline.Task = (*validatorsParserTask)(nil)

	errUnexpectedTxDataFormat    = errors.New("unexpected data format for transaction")
	errUnexpectedEventDataFormat = errors.New("unexpected event data format")

	zero big.Int
)

func NewBlockParserTask() *blockParserTask {
	return &blockParserTask{}
}

type blockParserTask struct{}

type ParsedBlockData struct {
	ExtrinsicsCount         int64
	UnsignedExtrinsicsCount int64
	SignedExtrinsicsCount   int64
}

func (t *blockParserTask) GetName() string {
	return BlockParserTaskName
}

func (t *blockParserTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

	fetchedBlock := payload.RawBlock

	parsedBlockData := ParsedBlockData{
		ExtrinsicsCount: int64(len(fetchedBlock.GetExtrinsics())),
	}

	for _, extrinsic := range fetchedBlock.GetExtrinsics() {
		if extrinsic.IsSignedTransaction {
			parsedBlockData.SignedExtrinsicsCount += 1
		} else {
			parsedBlockData.UnsignedExtrinsicsCount += 1
		}
	}
	payload.ParsedBlock = parsedBlockData
	return nil
}

func NewValidatorsParserTask(cfg *config.Config, accountClient client.AccountClient, rewardsDb store.Rewards, syncablesDb store.Syncables, validatorDb store.ValidatorEraSeq) *validatorsParserTask {
	return &validatorsParserTask{
		cfg:           cfg,
		accountClient: accountClient,
		rewardsDb:     rewardsDb,
		syncablesDb:   syncablesDb,
		validatorDb:   validatorDb,
	}
}

type validatorsParserTask struct {
	cfg *config.Config

	accountClient client.AccountClient

	rewardsDb   store.Rewards
	syncablesDb store.Syncables
	validatorDb store.ValidatorEraSeq
}

// ParsedValidatorsData normalized validator data
type ParsedValidatorsData map[string]parsedValidator
type parsedValidator struct {
	Staking       *stakingpb.Validator
	Performance   *validatorperformancepb.Validator
	DisplayName   string
	parsedRewards parsedRewards
}

func (t *validatorsParserTask) GetName() string {
	return ValidatorsParserTaskName
}

func (t *validatorsParserTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

	rawStakingState := payload.RawStaking
	rawValidatorPerformances := payload.RawValidatorPerformance

	var err error
	var c RewardsCalculator
	if payload.RawStaking.GetTotalRewardPoints() != 0 && payload.RawStaking.GetTotalRewardPayout() != "" {
		c, err = newRewardsCalulator(payload.RawStaking.GetTotalRewardPoints(), payload.RawStaking.GetTotalRewardPayout())
		if err != nil {
			return err
		}
	}

	parsedValidatorsData := make(ParsedValidatorsData)

	// Get validator staking info
	for _, rawValidatorStakingInfo := range rawStakingState.GetValidators() {
		stashAccount := rawValidatorStakingInfo.GetStashAccount()

		identity, err := t.accountClient.GetIdentity(stashAccount)
		if err != nil {
			return err
		}

		var parsedData parsedValidator
		parsedData, _ = parsedValidatorsData[stashAccount]

		parsedData.Staking = rawValidatorStakingInfo
		parsedData.DisplayName = strings.TrimSpace(identity.GetIdentity().GetDisplayName())

		if c != nil {
			parsedRewards := t.getUnclaimedRewardData(c, rawValidatorStakingInfo)
			parsedRewards.Era = payload.Syncable.Era
			parsedData.parsedRewards = parsedRewards
		}
		parsedValidatorsData[stashAccount] = parsedData
	}

	// Get validator performance info
	for _, rawValidatorPerf := range rawValidatorPerformances {
		stashAccount := rawValidatorPerf.GetStashAccount()

		var parsedData parsedValidator
		parsedData, _ = parsedValidatorsData[stashAccount]

		parsedData.Performance = rawValidatorPerf
		parsedValidatorsData[stashAccount] = parsedData
	}

	payload.ParsedValidators = parsedValidatorsData
	return nil
}

func (t *validatorsParserTask) getUnclaimedRewardData(calc RewardsCalculator, rawValidator *stakingpb.Validator) parsedRewards {
	data := parsedRewards{}

	commPayout, leftoverPayout := calc.commissionPayout(rawValidator.GetRewardPoints(), rawValidator.GetCommission())

	if commPayout.Cmp(&zero) > 0 {
		data.Commission = commPayout.String()
	}
	if leftoverPayout.Cmp(&zero) <= 0 { // TODO check 100% commission
		return data
	}

	validatorStake := *big.NewInt(rawValidator.GetTotalStake())
	rewardPayout := calc.nominatorPayout(leftoverPayout, *big.NewInt(rawValidator.GetOwnStake()), validatorStake)
	if rewardPayout.Cmp(&zero) > 0 {
		data.Reward = rewardPayout.String()
	}

	for _, n := range rawValidator.GetStakers() {
		if !n.GetIsRewardEligible() {
			continue
		}

		amount := calc.nominatorPayout(leftoverPayout, *big.NewInt(n.GetStake()), validatorStake)
		if amount.Cmp(&zero) <= 0 {
			continue
		}

		reward := stakerReward{
			Amount: amount.String(),
			Stash:  n.GetStashAccount(),
		}
		data.StakerRewards = append(data.StakerRewards, reward)
	}
	return data
}

func NewTransactionParserTask(cfg *config.Config, accountClient client.AccountClient, rewardsDb store.Rewards, syncablesDb store.Syncables, validatorDb store.ValidatorEraSeq) *transactionParserTask {
	return &transactionParserTask{
		cfg:           cfg,
		accountClient: accountClient,
		rewardsDb:     rewardsDb,
		syncablesDb:   syncablesDb,
		validatorDb:   validatorDb,
	}
}

type transactionParserTask struct {
	cfg *config.Config

	accountClient client.AccountClient

	rewardsDb   store.Rewards
	syncablesDb store.Syncables
	validatorDb store.ValidatorEraSeq
}
type parsedRewards struct {
	Commission          string
	Reward              string
	RewardAndCommission string
	StakerRewards       []stakerReward
	IsClaimed           bool
	Era                 int64
}

type stakerReward struct {
	Stash  string
	Amount string
}

func (t *transactionParserTask) GetName() string {
	return TransactionParserTaskName
}

func (t *transactionParserTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))

	for _, tx := range payload.RawTransactions {
		var claims []RewardsClaim
		var err error

		if tx.GetSection() == sectionStaking && tx.GetMethod() == txMethodPayoutStakers {
			claim, err := getRewardsClaimFromPayoutStakersTx(tx.GetArgs())
			if err != nil {
				return err
			}
			claims = append(claims, claim)
		} else if tx.GetSection() == sectionUtility && (tx.GetMethod() == txMethodBatch || tx.GetMethod() == txMethodBatchAll) {
			claims, err = getRewardsClaimFromBatchTx(tx.GetArgs())
			if err != nil {
				return err
			}
		}

		shouldParseEvents, err := t.addRewardsClaimed(tx, claims, &payload.RewardsClaimed)

		if !shouldParseEvents || len(payload.RawEvents) == 0 {
			return nil
		}

		rewards, err := t.getRewardsFromEvents(tx.GetExtrinsicIndex(), claims, payload.RawEvents)
		if err != nil {
			return err
		}

		payload.RewardEraSequences = append(payload.RewardEraSequences, rewards...)
	}

	return nil
}

func (t *transactionParserTask) getRewardsFromEvents(txIdx int64, claims []RewardsClaim, events []*eventpb.Event) ([]model.RewardEraSeq, error) {
	var rewards []model.RewardEraSeq

	if len(events) == 0 {
		return rewards, nil
	}

	var idx int
	var nextVald string
	for i, claim := range claims {
		if i < len(claims)-1 {
			nextVald = claims[i+1].ValidatorStash
		} else {
			nextVald = ""
		}

		eraSeq, err := t.getEraSeq(claim.Era)
		if err != nil {
			return rewards, err
		}

		extractedRewards, ranged, err := t.extractRewardsForClaimFromEvents(claim, eraSeq, txIdx, nextVald, events[idx:])
		if err != nil {
			return rewards, err
		}
		idx += ranged

		count, err := t.rewardsDb.GetCount(claim.ValidatorStash, claim.Era)
		if err != nil {
			return rewards, err
		}

		// if already exists in db, then claim was added to RewardsClaimed, so don't add parsedRewards
		if count != 0 {
			continue
		}
		rewards = append(rewards, extractedRewards...)
	}
	return rewards, nil
}

func (t *transactionParserTask) getEraSeq(era int64) (model.EraSequence, error) {
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

	lastSyncableInEra, err := t.syncablesDb.FindLastInEra(era)
	if err != nil {
		return model.EraSequence{}, err
	}

	return model.EraSequence{
		Era:         era,
		StartHeight: firstHeightInEra,
		EndHeight:   lastSyncableInEra.Height,
		Time:        lastSyncableInEra.Time,
	}, nil
}

func (t *transactionParserTask) addRewardsClaimed(tx *transactionpb.Annotated, claims []RewardsClaim, rewardClaims *[]RewardsClaim) (shouldParseEvents bool, err error) {
	var count int64
	for _, claim := range claims {
		count, err = t.rewardsDb.GetCount(claim.ValidatorStash, claim.Era)
		if err != nil {
			return
		}
		if count == 0 {
			// doesn't exist in db, so need to create reward seqs from raw events
			shouldParseEvents = true
			continue
		}
		// exists in db, so can mark all claimed
		*rewardClaims = append(*rewardClaims, claim)
	}
	return
}

func (t *transactionParserTask) extractRewardsForClaimFromEvents(claim RewardsClaim, eraSeq model.EraSequence, txIdx int64, nextVald string, events []*eventpb.Event) ([]model.RewardEraSeq, int, error) {
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

func (t *transactionParserTask) getStashAndAmountFromData(event *eventpb.Event) (stash string, amount types.Quantity, err error) {
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
	if err != nil {
		return RewardsClaim{}, err
	}

	return getStashAndEraFromPayoutArgs(data)
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

type batchArgs struct {
	Method  string   `json:"method"`
	Section string   `json:"section"`
	Args    []string `json:"args"`
}

func getRewardsClaimFromBatchTx(args string) ([]RewardsClaim, error) {
	var resp []RewardsClaim
	var data [][]batchArgs

	err := json.Unmarshal([]byte(args), &data)
	if err != nil {
		return resp, err
	}

	if len(data) != 1 {
		return resp, errUnexpectedTxDataFormat
	}

	for _, batch := range data[0] {
		if batch.Method != "payoutStakers" {
			continue
		}
		r, err := getStashAndEraFromPayoutArgs(batch.Args)
		if err != nil {
			return resp, err
		}
		resp = append(resp, r)
	}

	return resp, nil
}
