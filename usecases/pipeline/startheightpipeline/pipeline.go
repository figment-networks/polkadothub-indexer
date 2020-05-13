package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/repos/syncablerepo"
	"github.com/figment-networks/polkadothub-indexer/utils/iterators"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type Pipeline interface {
	Start(ctx context.Context, iter pipeline.Iterator) Results
}

type processingPipeline struct {
	config   Config
	pipeline *pipeline.Pipeline
}

type Config struct {
	Report         report.Model
	SyncableDbRepo syncablerepo.DbRepo
}

type Results struct {
	SuccessCount int64
	ErrorCount   int64
	Error        *string
	Details      []byte
}

func NewPipeline(
	config Config,
	stages ...pipeline.StageRunner,
) Pipeline {
	// Assemble pipeline
	p := pipeline.New(stages...)

	return &processingPipeline{
		config:   config,
		pipeline: p,
	}
}

// Process block until the link iterator is exhausted, an error occurs or the
// context is cancelled.
func (c *processingPipeline) Start(ctx context.Context, iter pipeline.Iterator) Results {
	i := iter.(*iterators.HeightIterator)
	source := NewBlockSource(i)
	sink := NewSink(c.config.SyncableDbRepo, c.config.Report)
	err := c.pipeline.Process(ctx, source, sink)
	//TODO: Add details to results
	r := Results{
		SuccessCount: sink.Count(),
		ErrorCount:   i.Length() - sink.Count(),
	}
	if err != nil {
		errMsg := err.Error()
		r.Error = &errMsg
	}
	return r
}
