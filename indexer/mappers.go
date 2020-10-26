package indexer

import (
	"encoding/json"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/pkg/errors"
)

var (
	ErrBlockSequenceNotValid            = errors.New("block sequence not valid")
	ErrValidatorSessionSequenceNotValid = errors.New("validator session sequence not valid")
	ErrValidatorEraSequenceNotValid     = errors.New("validator era sequence not valid")
	ErrAccountEraSequenceNotValid       = errors.New("account era sequence not valid")
	ErrEventSequenceNotValid            = errors.New("event sequence not valid")
	ErrTransactionSequenceNotValid      = errors.New("transaction sequence not valid")
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
				Time:        syncable.Time,
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
				Time:        syncable.Time,
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

func ToEventSequence(syncable *model.Syncable, rawEvents []*eventpb.Event) ([]model.EventSeq, error) {
	var events []model.EventSeq
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

		events = append(events, e)
	}
	return events, nil
}

func ToAccountEraSequence(syncable *model.Syncable, firstHeight int64, rawStakingValidator *stakingpb.Validator) ([]model.AccountEraSeq, error) {
	var accountEraSeqs []model.AccountEraSeq

	eraSeq := &model.EraSequence{
		Era:         syncable.Era,
		StartHeight: firstHeight,
		EndHeight:   syncable.Height,
		Time:        syncable.Time,
	}
	validatorStashAccount := rawStakingValidator.GetStashAccount()
	validatorControllerAccount := rawStakingValidator.GetControllerAccount()

	e := model.AccountEraSeq{
		EraSequence: eraSeq,

		StashAccount:               validatorStashAccount,
		ControllerAccount:          validatorControllerAccount,
		ValidatorStashAccount:      validatorStashAccount,
		ValidatorControllerAccount: validatorControllerAccount,
		Stake:                      types.NewQuantityFromInt64(rawStakingValidator.GetOwnStake()),
	}

	if !e.Valid() {
		return nil, ErrAccountEraSequenceNotValid
	}

	accountEraSeqs = append(accountEraSeqs, e)

	for _, rawStakingNominator := range rawStakingValidator.GetStakers() {
		e = model.AccountEraSeq{
			EraSequence: eraSeq,

			StashAccount:               rawStakingNominator.GetStashAccount(),
			ControllerAccount:          rawStakingNominator.GetControllerAccount(),
			ValidatorStashAccount:      validatorStashAccount,
			ValidatorControllerAccount: validatorControllerAccount,
			Stake:                      types.NewQuantityFromInt64(rawStakingNominator.GetStake()),
		}

		if !e.Valid() {
			return nil, ErrAccountEraSequenceNotValid
		}

		accountEraSeqs = append(accountEraSeqs, e)
	}

	return accountEraSeqs, nil
}

func ToTransactionSequence(syncable *model.Syncable, rawTransactions []*transactionpb.Annotated) ([]model.TransactionSeq, error) {
	var transactions []model.TransactionSeq

	for _, rawTx := range rawTransactions {
		if !rawTx.GetIsSigned() {
			continue
		}

		tx := model.TransactionSeq{
			Sequence: &model.Sequence{
				Height: syncable.Height,
				Time:   syncable.Time,
			},

			Index:   rawTx.GetExtrinsicIndex(),
			Hash:    rawTx.GetHash(),
			Section: rawTx.GetSection(),
			Method:  rawTx.GetMethod(),
		}

		if !tx.Valid() {
			return nil, ErrTransactionSequenceNotValid
		}

		transactions = append(transactions, tx)
	}
	return transactions, nil
}
