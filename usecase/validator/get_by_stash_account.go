package validator

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getByStashAccountUseCase struct {
	accountEraSeqDb store.AccountEraSeq
	validatorDb     store.Validators
}

func NewGetByStashAccountUseCase(accountEraSeqDb store.AccountEraSeq, validatorDb store.Validators) *getByStashAccountUseCase {
	return &getByStashAccountUseCase{
		accountEraSeqDb: accountEraSeqDb,
		validatorDb:     validatorDb,
	}
}

func (uc *getByStashAccountUseCase) Execute(stashAccount string, sessionsLimit int64, erasLimit int64) (*AggDetailsView, error) {
	validatorAggs, err := uc.validatorDb.FindAggByStashAccount(stashAccount)
	if err != nil {
		return nil, err
	}

	sessionSequences, err := uc.getSessionSequences(stashAccount, sessionsLimit)
	if err != nil {
		return nil, err
	}

	eraSequences, err := uc.getEraSequences(stashAccount, erasLimit)
	if err != nil {
		return nil, err
	}

	accountEraSeqs, err := uc.accountEraSeqDb.FindLastByValidatorStashAccount(stashAccount)
	if err != nil {
		return nil, err
	}

	return ToAggDetailsView(validatorAggs, sessionSequences, eraSequences, accountEraSeqs), nil
}

func (uc *getByStashAccountUseCase) getSessionSequences(stashAccount string, sequencesLimit int64) ([]model.ValidatorSessionSeq, error) {
	var sequences []model.ValidatorSessionSeq
	var err error
	if sequencesLimit > 0 {
		sequences, err = uc.validatorDb.FindLastSessionSeqByStashAccount(stashAccount, sequencesLimit)
		if err != nil {
			return nil, err
		}
	}
	return sequences, nil
}

func (uc *getByStashAccountUseCase) getEraSequences(stashAccount string, sequencesLimit int64) ([]model.ValidatorEraSeq, error) {
	var sequences []model.ValidatorEraSeq
	var err error
	if sequencesLimit > 0 {
		sequences, err = uc.validatorDb.FindLastEraSeqByStashAccount(stashAccount, sequencesLimit)
		if err != nil {
			return nil, err
		}
	}
	return sequences, nil
}
