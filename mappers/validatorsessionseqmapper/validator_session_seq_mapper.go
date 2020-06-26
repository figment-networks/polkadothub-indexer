package validatorsessionseqmapper

import (
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/models/validatorsessionseq"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
)

func ToSequence(blockSyncable syncable.Model, validatorPerformanceSyncable syncable.Model) ([]validatorsessionseq.Model, errors.ApplicationError) {
	blockData, err := syncablemapper.UnmarshalBlockData(blockSyncable.Data)
	if err != nil {
		return nil, err
	}
	validatorPerfData, err := syncablemapper.UnmarshalValidatorPerformanceData(validatorPerformanceSyncable.Data)
	if err != nil {
		return nil, err
	}

	var validators []validatorsessionseq.Model
	for _, rawValidator := range validatorPerfData.GetValidators() {
		acc := validatorsessionseq.Model{
			SessionSequence: &shared.SessionSequence{
				ChainUid:       blockData.GetChain(),
				SpecVersionUid: blockData.GetSpecVersion(),
				Session:        blockData.GetSession(),
			},

			StashAccount: rawValidator.GetStashAccount(),
			Online: rawValidator.GetOnline(),
		}

		if !acc.Valid() {
			return nil, errors.NewErrorFromMessage("validator session sequence not valid", errors.NotValid)
		}

		validators = append(validators, acc)
	}
	return validators, nil
}

