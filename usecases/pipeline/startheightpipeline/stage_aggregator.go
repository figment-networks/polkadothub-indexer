package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type Aggregator interface {
	Process(context.Context, pipeline.Payload) (pipeline.Payload, error)
}

type aggregator struct {
	previous *payload
	payloads []*payload
}

func NewAggregator() *aggregator {
	return &aggregator{
	}
}

func (a *aggregator) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*payload)

	return payload, nil
}

