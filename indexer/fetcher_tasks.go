package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-proxy/grpc/height/heightpb"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	FetcherTaskName = "Fetcher"
)

type HeightMeta struct {
	Height        int64
	Time          types.Time
	SpecVersion   string
	ChainUID      string
	Session       int64
	Era           int64
	LastInSession bool
	LastInEra     bool
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
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageFetcher, t.GetName(), payload.CurrentHeight))

	resp, err := t.client.GetAll(payload.CurrentHeight)
	if err != nil {
		return err
	}

	logger.DebugJSON(resp,
		logger.Field("process", "pipeline"),
		logger.Field("stage", "fetcher"),
		logger.Field("request", "getAll"),
		logger.Field("height", payload.CurrentHeight),
	)

	payload.RawBlock = resp.GetBlock().GetBlock()
	payload.RawValidatorPerformance = resp.GetValidatorPerformance().GetValidators()
	payload.RawStaking = resp.GetStaking().GetStaking()
	payload.RawEvents = resp.GetEvent().GetEvents()
	payload.RawTransactions = resp.GetTransaction().GetTransactions()

	meta := resp.GetChain()
	payload.HeightMeta = HeightMeta{
		Height:        payload.CurrentHeight,
		Time:          *types.NewTimeFromTimestamp(*meta.GetTime()),
		ChainUID:      meta.GetChain(),
		SpecVersion:   meta.GetSpecVersion(),
		Session:       meta.GetSession(),
		Era:           meta.GetEra(),
		LastInSession: meta.GetLastInSession(),
		LastInEra:     meta.GetLastInEra(),
	}

	return nil
}
