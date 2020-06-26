package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"time"
)

const (
	BlockSeqCreatorTaskName               = "BlockSeqCreator"
	ValidatorSeqCreatorTaskName           = "ValidatorSeqCreator"
	TransactionSeqCreatorTaskName         = "TransactionSeqCreator"
	StakingSeqCreatorTaskName             = "StakingSeqCreator"
	DelegationSeqCreatorTaskName          = "DelegationSeqCreator"
	DebondingDelegationSeqCreatorTaskName = "DebondingDelegationSeqCreator"
)

var (
	_ pipeline.Task = (*blockSeqCreatorTask)(nil)
	//_ pipeline.Task = (*validatorSeqCreatorTask)(nil)
	//_ pipeline.Task = (*transactionSeqCreatorTask)(nil)
	//_ pipeline.Task = (*stakingSeqCreatorTask)(nil)
	//_ pipeline.Task = (*delegationSeqCreatorTask)(nil)
	//_ pipeline.Task = (*debondingDelegationSeqCreatorTask)(nil)
)

func NewBlockSeqCreatorTask(db *store.Store) *blockSeqCreatorTask {
	return &blockSeqCreatorTask{
		db: db,
	}
}

type blockSeqCreatorTask struct {
	db *store.Store
}

func (t *blockSeqCreatorTask) GetName() string {
	return BlockSeqCreatorTaskName
}

func (t *blockSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))

	rawBlockSeq, err := BlockToSequence(payload.Syncable, payload.RawBlock, payload.ParsedBlock)
	if err != nil {
		return err
	}

	blockSeq, err := t.db.BlockSeq.FindByHeight(payload.CurrentHeight)
	if err != nil {
		if err == store.ErrNotFound {
			payload.NewBlockSequence = rawBlockSeq
			return nil
		} else {
			return err
		}
	}

	blockSeq.Update(*rawBlockSeq)
	payload.UpdatedBlockSequence = blockSeq

	return nil
}
//
//func NewValidatorSeqCreatorTask(db *store.Store) *validatorSeqCreatorTask {
//	return &validatorSeqCreatorTask{
//		db: db,
//	}
//}
//
//type validatorSeqCreatorTask struct {
//	db *store.Store
//}
//
//func (t *validatorSeqCreatorTask) GetName() string {
//	return ValidatorSeqCreatorTaskName
//}
//
//func (t *validatorSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
//
//	rawValidatorSeqs, err := ValidatorToSequence(payload.Syncable, payload.RawValidators, payload.ParsedValidators)
//	if err != nil {
//		return err
//	}
//
//	var newValidatorSeqs []model.ValidatorSeq
//	var updatedValidatorSeqs []model.ValidatorSeq
//	for _, rawValidatorSeq := range rawValidatorSeqs {
//		validatorSeq, err := t.db.ValidatorSeq.FindByHeightAndEntityUID(payload.CurrentHeight, rawValidatorSeq.EntityUID)
//		if err != nil {
//			if err == store.ErrNotFound {
//				newValidatorSeqs = append(newValidatorSeqs, rawValidatorSeq)
//				continue
//			} else {
//				return err
//			}
//		}
//
//		validatorSeq.Update(rawValidatorSeq)
//		updatedValidatorSeqs = append(updatedValidatorSeqs, *validatorSeq)
//	}
//
//	payload.NewValidatorSequences = newValidatorSeqs
//	payload.UpdatedValidatorSequences = updatedValidatorSeqs
//
//	return nil
//}
//
//func NewTransactionSeqCreatorTask(db *store.Store) *transactionSeqCreatorTask {
//	return &transactionSeqCreatorTask{
//		db: db,
//	}
//}
//
//type transactionSeqCreatorTask struct {
//	db *store.Store
//}
//
//func (t *transactionSeqCreatorTask) GetName() string {
//	return TransactionSeqCreatorTaskName
//}
//
//func (t *transactionSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
//
//	var res []model.TransactionSeq
//	sequenced, err := t.db.TransactionSeq.FindByHeight(payload.CurrentHeight)
//	if err != nil {
//		return err
//	}
//
//	toSequence, err := TransactionToSequence(payload.Syncable, payload.RawTransactions)
//	if err != nil {
//		return err
//	}
//
//	// Nothing to sequence
//	if len(toSequence) == 0 {
//		payload.TransactionSequences = res
//		return nil
//	}
//
//	// Everything sequenced and saved to persistence
//	if len(sequenced) == len(toSequence) {
//		payload.TransactionSequences = sequenced
//		return nil
//	}
//
//	isSequenced := func(vs model.TransactionSeq) bool {
//		for _, sv := range sequenced {
//			if sv.Equal(vs) {
//				return true
//			}
//		}
//		return false
//	}
//
//	for _, vs := range toSequence {
//		if !isSequenced(vs) {
//			if err := t.db.TransactionSeq.Create(&vs); err != nil {
//				return err
//			}
//		}
//		res = append(res, vs)
//	}
//	payload.TransactionSequences = res
//	return nil
//}
//
//func NewStakingSeqCreatorTask(db *store.Store) *stakingSeqCreatorTask {
//	return &stakingSeqCreatorTask{
//		db: db,
//	}
//}
//
//type stakingSeqCreatorTask struct {
//	db *store.Store
//}
//
//func (t *stakingSeqCreatorTask) GetName() string {
//	return StakingSeqCreatorTaskName
//}
//
//func (t *stakingSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
//
//	sequenced, err := t.db.StakingSeq.FindByHeight(payload.CurrentHeight)
//	if err != nil {
//		if err == store.ErrNotFound {
//			toSequence, err := StakingToSequence(payload.Syncable, payload.RawState)
//			if err != nil {
//				return err
//			}
//			if err := t.db.StakingSeq.Create(toSequence); err != nil {
//				return err
//			}
//			payload.StakingSequence = toSequence
//			return nil
//		}
//		return err
//	}
//	payload.StakingSequence = sequenced
//	return nil
//}
//
//type delegationSeqCreatorTask struct {
//	db *store.Store
//}
//
//func (t *delegationSeqCreatorTask) GetName() string {
//	return DelegationSeqCreatorTaskName
//}
//
//func NewDelegationsSeqCreatorTask(db *store.Store) *delegationSeqCreatorTask {
//	return &delegationSeqCreatorTask{
//		db: db,
//	}
//}
//
//func (t *delegationSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
//
//	var res []model.DelegationSeq
//	sequenced, err := t.db.DelegationSeq.FindByHeight(payload.CurrentHeight)
//	if err != nil {
//		return err
//	}
//
//	toSequence, err := DelegationToSequence(payload.Syncable, payload.RawState)
//	if err != nil {
//		return err
//	}
//
//	// Nothing to sequence
//	if len(toSequence) == 0 {
//		payload.DelegationSequences = res
//		return nil
//	}
//
//	// Everything sequenced and saved to persistence
//	if len(sequenced) == len(toSequence) {
//		payload.DelegationSequences = sequenced
//		return nil
//	}
//
//	isSequenced := func(vs model.DelegationSeq) bool {
//		for _, sv := range sequenced {
//			if sv.Equal(vs) {
//				return true
//			}
//		}
//		return false
//	}
//
//	for _, vs := range toSequence {
//		if !isSequenced(vs) {
//			if err := t.db.DelegationSeq.Create(&vs); err != nil {
//				return err
//			}
//		}
//		res = append(res, vs)
//	}
//	payload.DelegationSequences = res
//	return nil
//}
//
//func NewDebondingDelegationsSeqCreatorTask(db *store.Store) *debondingDelegationSeqCreatorTask {
//	return &debondingDelegationSeqCreatorTask{
//		db: db,
//	}
//}
//
//type debondingDelegationSeqCreatorTask struct {
//	db *store.Store
//}
//
//func (t *debondingDelegationSeqCreatorTask) GetName() string {
//	return DebondingDelegationSeqCreatorTaskName
//}
//
//func (t *debondingDelegationSeqCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageSequencer, t.GetName(), payload.CurrentHeight))
//
//	var res []model.DebondingDelegationSeq
//	sequenced, err := t.db.DebondingDelegationSeq.FindByHeight(payload.CurrentHeight)
//	if err != nil {
//		return err
//	}
//
//	toSequence, err := DebondingDelegationToSequence(payload.Syncable, payload.RawState)
//	if err != nil {
//		return err
//	}
//
//	// Nothing to sequence
//	if len(toSequence) == 0 {
//		payload.DebondingDelegationSequences = res
//		return nil
//	}
//
//	// Everything sequenced and saved to persistence
//	if len(sequenced) == len(toSequence) {
//		payload.DebondingDelegationSequences = sequenced
//		return nil
//	}
//
//	isSequenced := func(vs model.DebondingDelegationSeq) bool {
//		for _, sv := range sequenced {
//			if sv.Equal(vs) {
//				return true
//			}
//		}
//		return false
//	}
//
//	for _, vs := range toSequence {
//		if !isSequenced(vs) {
//			if err := t.db.DebondingDelegationSeq.Create(&vs); err != nil {
//				return err
//			}
//		}
//		res = append(res, vs)
//	}
//	payload.DebondingDelegationSequences = res
//	return nil
//}
