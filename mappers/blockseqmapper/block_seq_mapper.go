package blockseqmapper

import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/blockseq"
	"github.com/figment-networks/polkadothub-indexer/models/extrinsicseq"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToSequence(blockSyncable syncable.Model) (*blockseq.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}

	var signedExtrinsicsCount int64 = 0
	var unsignedExtrinsicsCount int64 = 0
	for _, extrinsic := range blockData.Block.Extrinsics {
		if extrinsic.IsSignedTransaction {
			signedExtrinsicsCount += 1
		} else {
			unsignedExtrinsicsCount += 1
		}
	}

	e := &blockseq.Model{
		HeightSequence: &shared.HeightSequence{
			SpecVersionUid: blockData.SpecVersion,
			ChainUid: blockData.Chain,
			Height:  types.Height(blockSyncable.SequenceId),
			Time:    types.NewTimeFromTimestamp(*blockData.Block.Header.Time),
		},

		ParentHash: blockData.Block.Header.ParentHash,
		StateRoot: blockData.Block.Header.ParentHash,
		ExtrinsicsRoot: blockData.Block.Header.ParentHash,
		ExtrinsicsCount: int64(len(blockData.Block.Extrinsics)),
		SignedExtrinsicsCount: signedExtrinsicsCount,
		UnsignedExtrinsicsCount: unsignedExtrinsicsCount,
	}

	if !e.Valid() {
		return nil, errors.NewErrorFromMessage("block sequence not valid", errors.NotValid)
	}

	return e, nil
}

type DetailsView struct {
	*shared.Model
	*shared.HeightSequence

	ParentHash              string     `json:"parent_hash"`
	StateRoot               string     `json:"state_root"`
	ExtrinsicsRoot          string     `json:"extrinsics_root"`
	ExtrinsicsCount         int64      `json:"extrinsics_count"`

	Transactions []extrinsicseq.Model `json:"transactions"`
}

func ToDetailsView(m *blockseq.Model, s syncable.Model, ts []extrinsicseq.Model) (*DetailsView, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(s.Data)
	if err != nil {
		return nil, err
	}

	return &DetailsView{
		Model: m.Model,
		HeightSequence: m.HeightSequence,

		ParentHash: blockData.Block.Header.ParentHash,
		StateRoot: blockData.Block.Header.ParentHash,
		ExtrinsicsRoot: blockData.Block.Header.ParentHash,
		ExtrinsicsCount: int64(len(blockData.Block.Extrinsics)),

		Transactions: ts,
	}, nil
}
