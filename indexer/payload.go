package indexer

import (
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validator/validatorpb"
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

	// Fetcher stage
	HeightMeta              HeightMeta
	RawBlock                *blockpb.Block
	RawValidatorPerformance []*validatorperformancepb.Validator
	RawStaking              *stakingpb.Staking
	RawEvents               []*eventpb.Event
	RawTransactions         []*transactionpb.Annotated
	RawValidators           []*validatorpb.Validator

	// Syncer stage
	Syncable *model.Syncable

	// Parser stage
	ParsedBlock      ParsedBlockData
	ParsedValidators ParsedValidatorsData

	// Aggregator stage
	NewValidatorAggregates     []model.ValidatorAgg
	UpdatedValidatorAggregates []model.ValidatorAgg

	// Sequencer stage
	NewBlockSequence          *model.BlockSeq
	UpdatedBlockSequence      *model.BlockSeq
	ValidatorSequences        []model.ValidatorSeq
	ValidatorSessionSequences []model.ValidatorSessionSeq
	ValidatorEraSequences     []model.ValidatorEraSeq
	EventSequences            []model.EventSeq
	AccountEraSequences       []model.AccountEraSeq
	TransactionSequences      []model.TransactionSeq
	RewardEraSequences        []model.RewardEraSeq

	// Analyzer
	SystemEvents []model.SystemEvent
}

func (p *payload) MarkAsProcessed() {}
