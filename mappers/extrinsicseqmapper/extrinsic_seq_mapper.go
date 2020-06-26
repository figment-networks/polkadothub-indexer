package extrinsicseqmapper

import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/extrinsicseq"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToSequence(blockSyncable syncable.Model) ([]extrinsicseq.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}

	var extrinsics []extrinsicseq.Model
	for _, rawExtrinsic := range blockData.Block.Extrinsics {
		e := extrinsicseq.Model{
			HeightSequence: &shared.HeightSequence{
				ChainUid:       blockData.GetChain(),
				SpecVersionUid: blockData.GetSpecVersion(),
				Height:         types.Height(blockSyncable.SequenceId),
				Session:        blockData.GetSession(),
				Era:            blockData.GetEra(),
				Time:           types.NewTimeFromTimestamp(*blockData.GetBlock().GetHeader().GetTime()),
			},

			Index:     rawExtrinsic.ExtrinsicIndex,
			Signature: rawExtrinsic.Signature,
			Signer:    rawExtrinsic.Signer,
			Nonce:     rawExtrinsic.Nonce,
			Method:    rawExtrinsic.Method,
			Section:   rawExtrinsic.Section,
			Args:      rawExtrinsic.Args,
			IsSigned:  rawExtrinsic.IsSignedTransaction,
		}

		if !e.Valid() {
			return nil, errors.NewErrorFromMessage("transaction sequence not valid", errors.NotValid)
		}

		extrinsics = append(extrinsics, e)
	}
	return extrinsics, nil
}

type ListView struct {
	Items []extrinsicseq.Model `json:"items"`
}

func ToListView(ts []extrinsicseq.Model) *ListView {
	return &ListView{
		Items: ts,
	}
}
