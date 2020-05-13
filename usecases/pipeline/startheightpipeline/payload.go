package startheightpipeline

import (
	"github.com/figment-networks/polkadothub-indexer/models/blockseq"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/models/extrinsicseq"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
	"sync"
)

var (
	payloadPool = sync.Pool{
		New: func() interface{} {
			return new(payload)
		},
	}
)

type payload struct {
	StartHeight   types.Height
	EndHeight     types.Height
	CurrentHeight types.Height
	RetrievedAt   types.Time

	// Syncer stage
	BlockSyncable        *syncable.Model

	// Aggregate stage

	// Sequence stage
	BlockSequence      *blockseq.Model
	ExtrinsicSequences []extrinsicseq.Model
}

func (p *payload) Clone() pipeline.Payload {
	newP := payloadPool.Get().(*payload)

	newP.StartHeight = p.StartHeight
	newP.EndHeight = p.EndHeight
	newP.RetrievedAt = p.RetrievedAt

	newP.BlockSyncable = p.BlockSyncable

	newP.BlockSequence = p.BlockSequence
	newP.ExtrinsicSequences = append([]extrinsicseq.Model(nil), p.ExtrinsicSequences...)

	return newP
}

func (p *payload) MarkAsProcessed() {
	// Reset
	p.ExtrinsicSequences = p.ExtrinsicSequences[:0]

	payloadPool.Put(p)
}
