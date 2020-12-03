package indexer

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/metric"
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

func NewValidatorsParserTask(accountClient client.AccountClient) *validatorsParserTask {
	return &validatorsParserTask{
		accountClient: accountClient,
	}
}

type validatorsParserTask struct {
	accountClient client.AccountClient
}

// ParsedValidatorsData normalized validator data
type ParsedValidatorsData map[string]parsedValidator
type parsedValidator struct {
	Staking                *stakingpb.Validator
	Performance            *validatorperformancepb.Validator
	DisplayName            string
	UnclaimedCommission    string
	UnclaimedReward        string
	UnclaimedStakerRewards []stakerReward
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
			t.injectRewardData(c, &parsedData, rawValidatorStakingInfo)
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

	payload.ParsedValidators = parsedValidatorsData
	return nil
}

func (t *validatorsParserTask) injectRewardData(calc RewardsCalculator, data *parsedValidator, rawValidator *stakingpb.Validator) {
	commPayout, leftoverPayout := calc.commissionPayout(rawValidator.GetRewardPoints(), rawValidator.GetCommission())

	if commPayout.Cmp(&zero) > 0 {
		data.UnclaimedCommission = commPayout.String()
	}
	if leftoverPayout.Cmp(&zero) <= 0 { // TODO check 100% commission
		return
	}

	validatorStake := *big.NewInt(rawValidator.GetTotalStake())
	rewardPayout := calc.nominatorPayout(leftoverPayout, *big.NewInt(rawValidator.GetOwnStake()), validatorStake)
	if rewardPayout.Cmp(&zero) > 0 {
		data.UnclaimedReward = rewardPayout.String()
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
		data.UnclaimedStakerRewards = append(data.UnclaimedStakerRewards, reward)
	}
}
