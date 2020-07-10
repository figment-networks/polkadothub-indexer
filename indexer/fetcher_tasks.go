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
	BlockFetcherTaskName                = "BlockFetcher"
	StakingFetcherTaskName              = "StakingFetcher"
	ValidatorPerformanceFetcherTaskName = "ValidatorPerformanceFetcher"
	EventsFetcherTaskName               = "EventsFetcher"
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

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	block, err := t.client.Block.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(block.GetBlock(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "block"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawBlock = block.GetBlock()
	return nil
}

func NewValidatorPerformanceFetcherTask(client *client.Client) pipeline.Task {
	return &ValidatorPerformanceFetcherTask{
		client: client,
	}
}

type ValidatorPerformanceFetcherTask struct {
	client *client.Client
}

func (t *ValidatorPerformanceFetcherTask) GetName() string {
	return ValidatorPerformanceFetcherTaskName
}

func (t *ValidatorPerformanceFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in session [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	perf, err := t.client.ValidatorPerformance.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(perf.GetValidators(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "block"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawValidatorPerformance = perf.GetValidators()

	return nil
}

func NewStakingFetcherTask(client *client.Client) pipeline.Task {
	return &StakingFetcherTask{
		client: client,
	}
}

type StakingFetcherTask struct {
	client *client.Client
}

func (t *StakingFetcherTask) GetName() string {
	return StakingFetcherTaskName
}

func (t *StakingFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		logger.Info(fmt.Sprintf("indexer task skipped because height is not last in era [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	staking, err := t.client.Staking.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(staking.GetStaking(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "block"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawStaking = staking.GetStaking()

	return nil
}


func NewEventsFetcherTask(client *client.Client) pipeline.Task {
	return &EventsFetcherTask{
		client: client,
	}
}

type EventsFetcherTask struct {
	client *client.Client
}

func (t *EventsFetcherTask) GetName() string {
	return EventsFetcherTaskName
}

func (t *EventsFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	events, err := t.client.Event.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(events.GetEvents(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "block"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawEvents = events.GetEvents()

	return nil
}