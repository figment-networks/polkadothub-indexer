package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	BlockFetcherTaskName        = "BlockFetcher"
	StateFetcherTaskName        = "StateFetcher"
	StakingStateFetcherTaskName = "StakingStateFetcher"
	ValidatorFetcherTaskName    = "ValidatorFetcher"
	TransactionFetcherTaskName  = "TransactionFetcher"
)

func NewBlockFetcherTask(client *client.Client) pipeline.Task {
	return &BlockFetcherTask{
		client: client,
	}
}

type BlockFetcherTask struct {
	client *client.Client
}

func (t *BlockFetcherTask) GetName() string {
	return BlockFetcherTaskName
}

func (t *BlockFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)
	block, err := t.client.Block.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
	logger.DebugJSON(block.GetBlock(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "block"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawBlock = block.GetBlock()
	return nil
}
