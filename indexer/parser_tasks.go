package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/pkg/errors"
)

const (
	BlockParserTaskName      = "BlockParser"
	ValidatorsParserTaskName = "ValidatorsParser"

	txMethodPayoutStakers = "payoutStakers"
	eventMethodReward     = "Reward"
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

		parsedData, ok := parsedValidatorsData[stashAccount]
		if !ok {
			parsedData = parsedValidator{}
		}
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

		parsedData, ok := parsedValidatorsData[stashAccount]
		if !ok {
			parsedData = parsedValidator{}
		}
		parsedData.Performance = rawValidatorPerf
		parsedValidatorsData[stashAccount] = parsedData
	}

	// Get claimed validator rewards
	for _, tx := range payload.RawTransactions {
		if tx.GetMethod() != txMethodPayoutStakers || tx.GetSection() != sectionStaking {
			continue
		}

		validatorStash, era, err := getStashAndEraFromPayoutArgs(tx)
		if err != nil {
			return err
		}

		//check if already exists in db, if yes mark all claimed
		count, err := t.rewardsDb.GetCount(validatorStash, era)
		if err != nil {
			return err
		}

		if count == 0 {
			parsedData, ok := parsedValidatorsData[validatorStash]
			if !ok {
				parsedData = parsedValidator{}
			}

			// these are historical rewards whose unclaimed reward data is not in the database
			parsedRewards, err := t.getClaimedRewardDataFromEvents(validatorStash, era, payload.RawEvents)
			if err != nil {
				return err
			}
			parsedData.parsedRewards = parsedRewards
			parsedValidatorsData[validatorStash] = parsedData
			continue
		}

		payload.RewardsClaimed = append(payload.RewardsClaimed, RewardsClaim{
			Era:            era,
			ValidatorStash: validatorStash,
		})
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

func (t *validatorsParserTask) getClaimedRewardDataFromEvents(validatorStash string, era int64, events []*eventpb.Event) (parsedRewards, error) {
	if len(events) == 0 {
		return parsedRewards{}, nil
	}

	data := parsedRewards{IsClaimed: true, Era: era}
	var validatorRewardAndCommission types.Quantity

	for _, event := range events {
		if event.GetMethod() != eventMethodReward || event.GetSection() != sectionStaking {
			continue
		}

		stash, amount, err := t.getStashAndAmountFromData(event)
		if err != nil {
			return parsedRewards{}, err
		}

		if stash == validatorStash {
			validatorRewardAndCommission = amount
			continue
		}

		reward := stakerReward{
			Amount: amount.String(),
			Stash:  stash,
		}
		data.StakerRewards = append(data.StakerRewards, reward)
	}

	if validatorRewardAndCommission.Cmp(&zero) <= 0 {
		return data, nil
	}

	validator, err := t.validatorDb.FindByEraAndStashAccount(era, validatorStash)
	if err != nil {
		return parsedRewards{}, err
	}

	if validator.Commission == hundredpermill || validator.OwnStake.Cmp(&zero) <= 0 {
		data.Commission = validatorRewardAndCommission.String()
		return data, nil
	}

	if validator.Commission == 0 {
		data.Reward = validatorRewardAndCommission.String()
		return data, nil
	}

	data.RewardAndCommission = validatorRewardAndCommission.String()
	return data, nil
}

func (t *validatorsParserTask) getStashAndAmountFromData(event *eventpb.Event) (stash string, amount types.Quantity, err error) {
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

func getStashAndEraFromPayoutArgs(tx *transactionpb.Annotated) (validatorStash string, era int64, err error) {
	var data []string

	err = json.Unmarshal([]byte(tx.GetArgs()), &data)
	if err != nil {
		return validatorStash, era, err
	}

	if len(data) < 2 {
		return validatorStash, era, errUnexpectedTxDataFormat
	}
	validatorStash = data[0]
	era, err = strconv.ParseInt(data[1], 10, 64)
	return
}
