package indexer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
)

const (
	BlockParserTaskName      = "BlockParser"
	ValidatorsParserTaskName = "ValidatorsParser"
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

func (t *blockParserTask) GetName() string {
	return BlockParserTaskName
}

func (t *blockParserTask) Run(ctx context.Context, p pipeline.Payload) error {
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

		reward := stakerReward{
			Amount: amount.String(),
			Stash:  n.GetStashAccount(),
		}
		data.StakerRewards = append(data.StakerRewards, reward)
	}
	return data
}
