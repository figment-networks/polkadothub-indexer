package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/repos/syncablerepo"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type Sink interface {
	Consume(context.Context, pipeline.Payload) error
	Count() int64
}

type sink struct {
	syncableDbRepo syncablerepo.DbSaver
	report         report.Model

	count    int64
}

func NewSink(syncableDbRepo syncablerepo.DbSaver, report report.Model) Sink {
	return &sink{
		syncableDbRepo: syncableDbRepo,
		report:         report,
	}
}

func (s *sink) Consume(ctx context.Context, p pipeline.Payload) error {
	payload := p.(*payload)
	s.count = s.count + 1

	block := payload.BlockSyncable
	block.MarkProcessed(s.report.ID)
	s.syncableDbRepo.Save(block)

	return nil
}

func (s *sink) Count() int64 {
	return s.count
}
