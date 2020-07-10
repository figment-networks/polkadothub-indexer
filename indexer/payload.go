package indexer

import (
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
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
	RawBlock                *blockpb.Block
	RawValidatorPerformance []*validatorperformancepb.Validator
	RawStaking              *stakingpb.Staking
	RawEvents               []*eventpb.Event

	// Parser stage
	ParsedBlock      ParsedBlockData
	ParsedValidators ParsedValidatorsData

	// Aggregator stage
	NewValidatorAggregates     []model.ValidatorAgg
	UpdatedValidatorAggregates []model.ValidatorAgg

	// Sequencer stage
	NewBlockSequence                 *model.BlockSeq
	UpdatedBlockSequence             *model.BlockSeq
	NewValidatorSessionSequences     []model.ValidatorSessionSeq
	UpdatedValidatorSessionSequences []model.ValidatorSessionSeq
	NewValidatorEraSequences         []model.ValidatorEraSeq
	UpdatedValidatorEraSequences     []model.ValidatorEraSeq
	NewEventSequences                []model.EventSeq
	UpdatedEventSequences            []model.EventSeq
}

func (p *payload) MarkAsProcessed() {}
