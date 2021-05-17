package indexer

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-proxy/grpc/height/heightpb"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	FetcherTaskName                     = "Fetcher"
	ValidatorFetcherTaskName            = "ValidatorFetcher"
	ValidatorPerformanceFetcherTaskName = "ValidatorPerformanceFetcher"
)

type HeightMeta struct {
	Height          int64
	Time            types.Time
	SpecVersion     string
	ChainUID        string
	Session         int64
	Era             int64
	ActiveEra       int64
	LastInSession   bool
	LastInEra       bool
	LastInActiveEra bool
}

func NewFetcherTask(client FetcherClient) pipeline.Task {
	return &FetcherTask{
		client: client,
	}
}

type FetcherClient interface {
	GetAll(int64) (*heightpb.GetAllResponse, error)
}

type FetcherTask struct {
	client FetcherClient
}

func (t *FetcherTask) GetName() string {
	return FetcherTaskName
}

func (t *FetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	resp, err := t.client.GetAll(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(resp,
		logger.Field("process", "pipeline"),
		logger.Field("stage", pipeline.StageFetcher),
		logger.Field("request", "getAll"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawBlock = resp.GetBlock().GetBlock()
	payload.RawValidatorPerformance = resp.GetValidatorPerformance().GetValidators()
	payload.RawStaking = resp.GetStaking().GetStaking()
	payload.RawEvents = resp.GetEvent().GetEvents()

	meta := resp.GetChain()
	payload.HeightMeta = HeightMeta{
		Height:          payload.CurrentHeight,
		Time:            *types.NewTimeFromTimestamp(*meta.GetTime()),
		ChainUID:        meta.GetChain(),
		SpecVersion:     meta.GetSpecVersion(),
		Session:         meta.GetSession(),
		Era:             meta.GetEra(),
		ActiveEra:       meta.GetActiveEra(),
		LastInSession:   meta.GetLastInSession(),
		LastInEra:       meta.GetLastInEra(),
		LastInActiveEra: meta.GetLastInActiveEra(),
	}

	return nil
}

func NewValidatorFetcherTask(client client.ValidatorClient) pipeline.Task {
	return &ValidatorFetcherTask{
		client: client,
	}
}

type ValidatorFetcherTask struct {
	client client.ValidatorClient
}

func (t *ValidatorFetcherTask) GetName() string {
	return ValidatorFetcherTaskName
}

func (t *ValidatorFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	payload := p.(*payload)
	validators, err := t.client.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
	logger.DebugJSON(validators.GetValidators(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", pipeline.StageFetcher),
		logger.Field("request", "validator"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawValidators = validators.GetValidators()
	return nil
}

func NewValidatorPerformanceFetcherTask(client client.ValidatorPerformanceClient) pipeline.Task {
	return &ValidatorPerformanceFetcherTask{
		client: client,
	}
}

type ValidatorPerformanceFetcherTask struct {
	client client.ValidatorPerformanceClient
}

func (t *ValidatorPerformanceFetcherTask) GetName() string {
	return ValidatorPerformanceFetcherTaskName
}

func (t *ValidatorPerformanceFetcherTask) Run(ctx context.Context, p pipeline.Payload) error {
	payload := p.(*payload)
	validators, err := t.client.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))
	logger.DebugJSON(validators.GetValidators(),
		logger.Field("process", "pipeline"),
		logger.Field("stage", pipeline.StageFetcher),
		logger.Field("request", "validator"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawValidatorPerformance = validators.GetValidators()
	return nil
}
