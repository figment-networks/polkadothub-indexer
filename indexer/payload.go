package indexer

import (
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
)

var (
	_ pipeline.PayloadFactory = (*payloadFactory)(nil)
	_ pipeline.Payload        = (*payload)(nil)
)

func NewPayloadFactory() *payloadFactory {
	return &payloadFactory{}
}

type payloadFactory struct{}

func (pf *payloadFactory) GetPayload(currentHeight int64) pipeline.Payload {
	return &payload{
		CurrentHeight: currentHeight,
	}
}

type payload struct {
	CurrentHeight int64

	// Setup stage
	HeightMeta HeightMeta

	// Syncer stage
	Syncable *model.Syncable

	// Fetcher stage
	RawBlock        *blockpb.Block

	// Parser stage
	ParsedBlock      ParsedBlockData
	//ParsedValidators ParsedValidatorsData

	// Aggregator stage

	// Sequencer stage
	NewBlockSequence          *model.BlockSeq
	UpdatedBlockSequence      *model.BlockSeq
	//NewValidatorSequences     []model.ValidatorSeq
	//UpdatedValidatorSequences []model.ValidatorSeq

	//TransactionSequences         []model.TransactionSeq
}

func (p *payload) MarkAsProcessed() {}
