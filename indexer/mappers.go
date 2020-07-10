package indexer

import (
	"encoding/json"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/pkg/errors"
)

var (
	ErrBlockSequenceNotValid            = errors.New("block sequence not valid")
	ErrValidatorSessionSequenceNotValid = errors.New("validator session sequence not valid")
	ErrValidatorEraSequenceNotValid     = errors.New("validator era sequence not valid")
	ErrEventSequenceNotValid            = errors.New("event sequence not valid")
)

func ToBlockSequence(syncable *model.Syncable, rawBlock *blockpb.Block, blockParsedData ParsedBlockData) (*model.BlockSeq, error) {
	e := &model.BlockSeq{
		Sequence: &model.Sequence{
			Height: syncable.Height,
			Time:   syncable.Time,
		},

		ExtrinsicsCount:         blockParsedData.ExtrinsicsCount,
		SignedExtrinsicsCount:   blockParsedData.SignedExtrinsicsCount,
		UnsignedExtrinsicsCount: blockParsedData.UnsignedExtrinsicsCount,
	}

	if !e.Valid() {
		return nil, ErrBlockSequenceNotValid
	}

	return e, nil
}

func ToValidatorSessionSequence(syncable *model.Syncable, firstHeight int64, rawValidatorPerformance []*validatorperformancepb.Validator) ([]model.ValidatorSessionSeq, error) {
	var validators []model.ValidatorSessionSeq
	for _, rawValidator := range rawValidatorPerformance {
		e := model.ValidatorSessionSeq{
			SessionSequence: &model.SessionSequence{
				Session:     syncable.Session,
				StartHeight: firstHeight,
				EndHeight:   syncable.Height,
			},

			StashAccount: rawValidator.GetStashAccount(),
			Online:       rawValidator.GetOnline(),
		}

		if !e.Valid() {
			return nil, ErrValidatorSessionSequenceNotValid
		}

		validators = append(validators, e)
	}
	return validators, nil
}

func ToValidatorEraSequence(syncable *model.Syncable, firstHeight int64, rawStakingValidator []*stakingpb.Validator) ([]model.ValidatorEraSeq, error) {
	var validators []model.ValidatorEraSeq
	for i, rawValidator := range rawStakingValidator {
		e := model.ValidatorEraSeq{
			EraSequence: &model.EraSequence{
				Era:         syncable.Era,
				StartHeight: firstHeight,
				EndHeight:   syncable.Height,
			},

			Index:             int64(i),
			StashAccount:      rawValidator.GetStashAccount(),
			ControllerAccount: rawValidator.GetControllerAccount(),
			SessionAccounts:   rawValidator.GetSessionKeys(),
			TotalStake:        types.NewQuantityFromInt64(rawValidator.GetTotalStake()),
			OwnStake:          types.NewQuantityFromInt64(rawValidator.GetOwnStake()),
			StakersStake:      types.NewQuantityFromInt64(rawValidator.GetStakersStake()),
			RewardPoints:      rawValidator.GetRewardPoints(),
			Commission:        rawValidator.GetCommission(),
			StakersCount:      len(rawValidator.GetStakers()),
		}

		if !e.Valid() {
			return nil, ErrValidatorEraSequenceNotValid
		}

		validators = append(validators, e)
	}
	return validators, nil
}

func EventToSequence(syncable *model.Syncable, rawEvents []*eventpb.Event) ([]model.EventSeq, error) {
	var transactions []model.EventSeq
	for _, rawEvent := range rawEvents {
		eventData := rawEvent.GetData()
		eventDataJSON, err := json.Marshal(eventData)
		if err != nil {
			return nil, err
		}

		e := model.EventSeq{
			Sequence: &model.Sequence{
				Height: syncable.Height,
				Time:   syncable.Time,
			},

			Index:          rawEvent.GetIndex(),
			ExtrinsicIndex: rawEvent.GetExtrinsicIndex(),
			Section:        rawEvent.GetSection(),
			Phase:          rawEvent.GetPhase(),
			Method:         rawEvent.GetMethod(),
			Data:           types.Jsonb{RawMessage: eventDataJSON},
		}

		if !e.Valid() {
			return nil, ErrEventSequenceNotValid
		}

		transactions = append(transactions, e)
	}
	return transactions, nil
}
