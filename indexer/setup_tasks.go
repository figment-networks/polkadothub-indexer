package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	HeightMetaRetrieverTaskName = "HeightMetaRetriever"
)

func NewHeightMetaRetrieverTask(c *client.Client) pipeline.Task {
	return &heightMetaRetrieverTask{
		client: c,
	}
}

type heightMetaRetrieverTask struct {
	client *client.Client
}

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

func (t *heightMetaRetrieverTask) GetName() string {
	return HeightMetaRetrieverTaskName
}

func (t *heightMetaRetrieverTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSetup, t.GetName(), payload.CurrentHeight))

	meta, err := t.client.Chain.GeMetaByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

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
