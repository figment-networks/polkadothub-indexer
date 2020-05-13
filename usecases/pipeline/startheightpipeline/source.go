package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/utils/iterators"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type blockSource struct {
	iter      *iterators.HeightIterator
}

func NewBlockSource(i *iterators.HeightIterator) *blockSource {
	return &blockSource{
		iter: i,
	}
}

func (s *blockSource) Error() error {
	return s.iter.Error()
}

func (s *blockSource) Next(ctx context.Context) bool {
	return s.iter.Next()
}

func (s *blockSource) Payload() pipeline.Payload {
	p := payloadPool.Get().(*payload)

	p.CurrentHeight = s.iter.Value()

	return p
}

