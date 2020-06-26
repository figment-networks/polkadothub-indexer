package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"time"
)

const (
	BlockParserTaskName      = "BlockParser"
	ValidatorsParserTaskName = "ValidatorsParser"
)

var (
	_ pipeline.Task = (*blockParserTask)(nil)
	//_ pipeline.Task = (*validatorsParserTask)(nil)
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
//
//func NewValidatorsParserTask() *validatorsParserTask {
//	return &validatorsParserTask{}
//}
//
//type validatorsParserTask struct{}
//
//type ParsedValidatorsData map[string]parsedValidator
//
//type parsedValidator struct {
//	Proposed             bool
//	PrecommitValidated   *bool
//	PrecommitBlockIdFlag int64
//	PrecommitIndex       int64
//	TotalShares          types.Quantity
//}
//
//func (t *validatorsParserTask) GetName() string {
//	return ValidatorsParserTaskName
//}
//
//func (t *validatorsParserTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageParser, t.GetName(), payload.CurrentHeight))
//
//	fetchedValidators := payload.RawValidators
//	fetchedBlock := payload.RawBlock
//	fetchedStakingState := payload.RawStakingState
//
//	parsedData := make(ParsedValidatorsData)
//	for i, rv := range fetchedValidators {
//		key := rv.Node.EntityId
//		calculatedData := parsedValidator{}
//
//		// Get precommit data
//		votes := fetchedBlock.GetLastCommit().GetVotes()
//		var index int64
//		// 1 = Not validated
//		// 2 = Validated
//		// 3 = Validated nil
//		var blockIdFlag int64
//		var validated *bool
//		if len(votes) > 0 {
//			// Account for situation when there is more validators than precommits
//			// It means that last x validators did not have chance to vote. In that case set validated to null.
//			if i > len(votes)-1 {
//				index = int64(i)
//				blockIdFlag = 3
//			} else {
//				precommit := votes[i]
//				isValidated := precommit.BlockIdFlag == 2
//				validated = &isValidated
//				index = precommit.ValidatorIndex
//				blockIdFlag = precommit.BlockIdFlag
//			}
//		} else {
//			index = int64(i)
//			blockIdFlag = 3
//		}
//
//		calculatedData.PrecommitValidated = validated
//		calculatedData.PrecommitIndex = index
//		calculatedData.PrecommitBlockIdFlag = blockIdFlag
//
//		// Get proposed
//		calculatedData.Proposed = fetchedBlock.GetHeader().GetProposerAddress() == rv.Address
//
//		// Get total shares
//		delegations, ok := fetchedStakingState.GetDelegations()[rv.Node.EntityId]
//		totalShares := big.NewInt(0)
//		if ok {
//			for _, d := range delegations.Entries {
//				shares := types.NewQuantityFromBytes(d.Shares)
//				totalShares = totalShares.Add(totalShares, &shares.Int)
//			}
//		}
//		calculatedData.TotalShares = types.NewQuantity(totalShares)
//
//		parsedData[key] = calculatedData
//	}
//	payload.ParsedValidators = parsedData
//	return nil
//}
