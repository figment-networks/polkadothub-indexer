package indexer

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/pkg/errors"
)

var (
	ErrBlockSequenceNotValid            = errors.New("block sequence not valid")
	ErrValidatorSessionSequenceNotValid = errors.New("validator session sequence not valid")
	ErrValidatorEraSequenceNotValid     = errors.New("validator era sequence not valid")
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

			Index: int64(i),
			StashAccount: rawValidator.GetStashAccount(),
			ControllerAccount: rawValidator.GetControllerAccount(),
			SessionAccounts: rawValidator.GetSessionKeys(),
			TotalStake: types.NewQuantityFromInt64(rawValidator.GetTotalStake()),
			OwnStake: types.NewQuantityFromInt64(rawValidator.GetOwnStake()),
			StakersStake: types.NewQuantityFromInt64(rawValidator.GetStakersStake()),
			RewardPoints: rawValidator.GetRewardPoints(),
			Commission: rawValidator.GetCommission(),
			StakersCount: len(rawValidator.GetStakers()),
		}

		if !e.Valid() {
			return nil, ErrValidatorEraSequenceNotValid
		}

		validators = append(validators, e)
	}
	return validators, nil
}

//func TransactionToSequence(syncable *model.Syncable, rawTransactions []*transactionpb.Transaction) ([]model.TransactionSeq, error) {
//	var transactions []model.TransactionSeq
//	for _, rawTransaction := range rawTransactions {
//		e := model.TransactionSeq{
//			Sequence: &model.Sequence{
//				Height: syncable.Height,
//				Time:   syncable.Time,
//			},
//
//			PublicKey: rawTransaction.GetPublicKey(),
//			Hash:      rawTransaction.GetHash(),
//			Nonce:     rawTransaction.GetNonce(),
//			Fee:       types.NewQuantityFromBytes(rawTransaction.GetFee()),
//			GasLimit:  rawTransaction.GetGasLimit(),
//			GasPrice:  types.NewQuantityFromBytes(rawTransaction.GetGasPrice()),
//			Method:    rawTransaction.GetMethod(),
//		}
//
//		if !e.Valid() {
//			return nil, errors.New("transaction sequence not valid")
//		}
//
//		transactions = append(transactions, e)
//	}
//	return transactions, nil
//}
//
//func StakingToSequence(syncable *model.Syncable, rawState *statepb.State) (*model.StakingSeq, error) {
//	e := &model.StakingSeq{
//		Sequence: &model.Sequence{
//			Height: syncable.Height,
//			Time:   syncable.Time,
//		},
//
//		TotalSupply:         types.NewQuantityFromBytes(rawState.GetStaking().GetTotalSupply()),
//		CommonPool:          types.NewQuantityFromBytes(rawState.GetStaking().GetCommonPool()),
//		DebondingInterval:   rawState.GetStaking().GetParameters().GetDebondingInterval(),
//		MinDelegationAmount: types.NewQuantityFromBytes(rawState.GetStaking().GetParameters().GetMinDelegationAmount()),
//	}
//
//	if !e.Valid() {
//		return nil, errors.New("staking sequence not valid")
//	}
//
//	return e, nil
//}
//
//func DelegationToSequence(syncable *model.Syncable, rawState *statepb.State) ([]model.DelegationSeq, error) {
//	var delegations []model.DelegationSeq
//	for validatorUID, delegationsMap := range rawState.GetStaking().GetDelegations() {
//		for delegatorUID, info := range delegationsMap.GetEntries() {
//			acc := model.DelegationSeq{
//				Sequence: &model.Sequence{
//					Height: syncable.Height,
//					Time:   syncable.Time,
//				},
//
//				ValidatorUID: validatorUID,
//				DelegatorUID: delegatorUID,
//				Shares:       types.NewQuantityFromBytes(info.GetShares()),
//			}
//
//			if !acc.Valid() {
//				return nil, errors.New("delegation sequence not valid")
//			}
//
//			delegations = append(delegations, acc)
//		}
//	}
//	return delegations, nil
//}
//
//func DebondingDelegationToSequence(syncable *model.Syncable, rawState *statepb.State) ([]model.DebondingDelegationSeq, error) {
//	var delegations []model.DebondingDelegationSeq
//	for validatorUID, delegationsMap := range rawState.GetStaking().GetDebondingDelegations() {
//		for delegatorUID, infoArray := range delegationsMap.GetEntries() {
//			for _, delegation := range infoArray.GetDebondingDelegations() {
//				acc := model.DebondingDelegationSeq{
//					Sequence: &model.Sequence{
//						Height: syncable.Height,
//						Time:   syncable.Time,
//					},
//
//					ValidatorUID: validatorUID,
//					DelegatorUID: delegatorUID,
//					Shares:       types.NewQuantityFromBytes(delegation.GetShares()),
//					DebondEnd:    delegation.GetDebondEndTime(),
//				}
//
//				if !acc.Valid() {
//					return nil, errors.New("debonding delegation sequence not valid")
//				}
//
//				delegations = append(delegations, acc)
//			}
//		}
//	}
//	return delegations, nil
//}
