package validatorseqmapper


import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToSequence(blockSyncable syncable.Model, stakingSyncable syncable.Model, validatorPerformanceSyncable syncable.Model) ([]model.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}
	stakingData, err := syncablemapper.UnmarshalStakingData(stakingSyncable.Data)
	if err != nil {
		return nil, err
	}

	var validators []model.Model
	for i, rawValidator := range stakingData.GetStaking().GetValidators() {
		acc := model.Model{
			HeightSequence: &shared.HeightSequence{
				ChainUid:       blockData.GetChain(),
				SpecVersionUid: blockData.GetSpecVersion(),
				Height:         types.Height(blockData.GetBlock().GetHeader().GetHeight()),
				Session:        blockData.GetSession(),
				Era:            blockData.GetEra(),
				Time:           types.NewTimeFromTimestamp(*blockData.GetBlock().GetHeader().GetTime()),
			},

			Index: int64(i),
			StashAccount: rawValidator.GetStashAccount(),
			ControllerAccount: rawValidator.GetControllerAccount(),
			SessionAccount: "",
			TotalStake: types.NewQuantityFromInt64(rawValidator.GetTotalStake()),
			OwnStake: types.NewQuantityFromInt64(rawValidator.GetOwnStake()),
			RewardPoints: rawValidator.GetRewardPoints(),
			Commission: rawValidator.GetCommission(),
			NominatorsCount: int64(len(rawValidator.GetStakers())),
		}

		if !acc.Valid() {
			return nil, errors.NewErrorFromMessage("account aggregator not valid", errors.NotValid)
		}

		validators = append(validators, acc)
	}
	return validators, nil
}

