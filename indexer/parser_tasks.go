package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
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

func NewValidatorsParserTask() *validatorsParserTask {
	return &validatorsParserTask{}
}

type validatorsParserTask struct{}

// ParsedValidatorsData normalized validator data
type ParsedValidatorsData map[string]parsedValidator

type parsedValidator struct {
	Staking     *stakingpb.Validator
	Performance *validatorperformancepb.Validator
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

	parsedValidatorsData := make(ParsedValidatorsData)

	// Get validator staking info
	for _, rawValidatorStakingInfo := range rawStakingState.GetValidators() {
		stashAccount := rawValidatorStakingInfo.GetStashAccount()

		parsedData, ok := parsedValidatorsData[stashAccount]
		if !ok {
			parsedData = parsedValidator{}
		}
		parsedData.Staking = rawValidatorStakingInfo
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
