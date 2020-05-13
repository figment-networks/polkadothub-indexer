package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/mappers/blockseqmapper"
	"github.com/figment-networks/polkadothub-indexer/mappers/extrinsicseqmapper"
	"github.com/figment-networks/polkadothub-indexer/models/extrinsicseq"
	"github.com/figment-networks/polkadothub-indexer/repos/blockseqrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/extrinsicseqrepo"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type Sequencer interface {
	Process(context.Context, pipeline.Payload) (pipeline.Payload, error)
}

type sequencer struct {
	blockSeqDbRepo               blockseqrepo.DbRepo
	extrinsicSeqDbRepo         extrinsicseqrepo.DbRepo
}

func NewSequencer(
	blockSeqDbRepo blockseqrepo.DbRepo,
	extrinsicSeqDbRepo extrinsicseqrepo.DbRepo,
) Sequencer {
	return &sequencer{
		blockSeqDbRepo:               blockSeqDbRepo,
		extrinsicSeqDbRepo:         extrinsicSeqDbRepo,
	}
}

func (s *sequencer) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*payload)

	// Sequence block
	err := s.sequenceBlock(payload)
	if err != nil {
		return nil, err
	}

	// Sequence extrinsics
	err = s.sequenceExtrinsics(payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

/*************** Private ***************/

func (s *sequencer) sequenceBlock(payload *payload) errors.ApplicationError {
	sequenced, err := s.blockSeqDbRepo.GetByHeight(payload.CurrentHeight)
	if err != nil {
		if err.Status() == errors.NotFoundError {
			toSequence, err := blockseqmapper.ToSequence(*payload.BlockSyncable)
			if err != nil {
				return err
			}
			if err := s.blockSeqDbRepo.Create(toSequence); err != nil {
				return err
			}
			payload.BlockSequence = sequenced
			return nil
		}
		return err
	}
	payload.BlockSequence = sequenced
	return nil
}

func (s *sequencer) sequenceExtrinsics(payload *payload) errors.ApplicationError {
	var sequences []extrinsicseq.Model
	sequenced, err := s.extrinsicSeqDbRepo.GetByHeight(payload.CurrentHeight)
	if err != nil {
		return err
	}

	toSequence, err := extrinsicseqmapper.ToSequence(*payload.BlockSyncable)
	if err != nil {
		return err
	}

	// Nothing to sequence
	if len(toSequence) == 0 {
		return nil
	}

	// Everything sequenced and saved to persistence
	if len(sequenced) == len(toSequence) {
		return nil
	}

	isSequenced := func(vs extrinsicseq.Model) bool {
		for _, sv := range sequenced {
			if sv.Equal(vs) {
				return true
			}
		}
		return false
	}

	for _, vs := range toSequence {
		if !isSequenced(vs) {
			if err := s.extrinsicSeqDbRepo.Create(&vs); err != nil {
				return err
			}
		}
		sequences = append(sequences, vs)

	}
	payload.ExtrinsicSequences = sequences
	return nil
}
