package nominatoraggmapper

import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/nominatoragg"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToAggregate(blockSyncable syncable.Model, stakingSyncable syncable.Model) ([]nominatoragg.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}
	stakingData, err := syncablemapper.UnmarshalStakingData(stakingSyncable.Data)
	if err != nil {
		return nil, err
	}

	var nominators []nominatoragg.Model
	for i, rawValidator := range stakingData.GetStaking().GetValidators() {
		for j, rawNominator := range rawValidator.GetStakers() {
			e := nominatoragg.Model{
				Aggregate: &shared.Aggregate{
					StartedAtHeight: types.Height(blockSyncable.SequenceId),
					StartedAt:       *types.NewTimeFromTimestamp(*blockData.GetBlock().GetHeader().GetTime()),
				},

				StashAccount:            rawNominator.GetStashAccount(),
				RecentControllerAccount: rawNominator.GetControllerAccount(),
				RecentIndex:             int64(j),
				RecentValidatorIndex:    int64(i),
				RecentStake:             types.NewQuantityFromInt64(rawNominator.GetStake()),
			}

			if !e.Valid() {
				return nil, errors.NewErrorFromMessage("nominator aggregate not valid", errors.NotValid)
			}

			nominators = append(nominators, e)
		}
	}
	return nominators, nil
}
