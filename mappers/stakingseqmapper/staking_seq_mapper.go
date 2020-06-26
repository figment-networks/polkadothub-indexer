package stakingseqmapper

import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/stakingseq"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToSequence(blockSyncable syncable.Model, stakingSyncable syncable.Model) (*stakingseq.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}
	stakingData, err := syncablemapper.UnmarshalStakingData(stakingSyncable.Data)
	if err != nil {
		return nil, err
	}

	e := &stakingseq.Model{
		HeightSequence: &shared.HeightSequence{
			ChainUid:       blockData.GetChain(),
			SpecVersionUid: blockData.GetSpecVersion(),
			Height:         types.Height(blockData.GetBlock().GetHeader().GetHeight()),
			Session:        blockData.GetSession(),
			Era:            blockData.GetEra(),
			Time:           types.NewTimeFromTimestamp(*blockData.GetBlock().GetHeader().GetTime()),
		},

		TotalStake:        types.NewQuantityFromInt64(stakingData.GetStaking().GetTotalStake()),
		TotalRewardPayout: types.NewQuantityFromInt64(stakingData.GetStaking().GetTotalRewardPayout()),
		TotalRewardPoints: types.NewQuantityFromInt64(stakingData.GetStaking().GetTotalRewardPoints()),
	}

	if !e.Valid() {
		return nil, errors.NewErrorFromMessage("block sequence not valid", errors.NotValid)
	}

	return e, nil
}
