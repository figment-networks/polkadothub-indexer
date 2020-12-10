package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	TaskNameEraSystemEventCreator     = "EraSystemEventCreator"
	TaskNameSessionSystemEventCreator = "SessionSystemEventCreator"
	TaskNameSystemEventCreator        = "SystemEventCreator"
)

var (
	ErrActiveBalanceOutsideOfRange = errors.New("active balance is outside of specified buckets")
	ErrCommissionOutsideOfRange    = errors.New("commission is outside of specified buckets")

	missedConsecutiveThreshold int64 = 1
)

// NewSystemEventCreatorTask creates system events
func NewSystemEventCreatorTask(cfg *config.Config, validatorSeqDb store.ValidatorSeq) *systemEventCreatorTask {
	return &systemEventCreatorTask{
		cfg: cfg,

		validatorSeqDb: validatorSeqDb,
	}
}

type systemEventCreatorTask struct {
	cfg *config.Config

	validatorSeqDb store.ValidatorSeq
}

func (t *systemEventCreatorTask) GetName() string {
	return TaskNameSystemEventCreator
}

func (t *systemEventCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", "Analyzer", t.GetName(), payload.CurrentHeight))

	prevHeightValidatorSeqs, err := t.getPrevHeightValidatorSequences(payload)
	if err != nil {
		return err
	}

	balanceChangeSystemEvents, err := t.getActiveBalanceChangeSystemEvents(payload.ValidatorSequences, prevHeightValidatorSeqs, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, balanceChangeSystemEvents...)
	return nil
}

// NewSessionSystemEventCreatorTask creates system events
func NewSessionSystemEventCreatorTask(cfg *config.Config, syncablesDb store.Syncables, systemEventDb store.SystemEvents, validatorSeqDb store.ValidatorSeq, validatorSessionSeqDb store.ValidatorSessionSeq,
) *sessionSystemEventCreatorTask {
	return &sessionSystemEventCreatorTask{
		cfg: cfg,

		syncablesDb:           syncablesDb,
		systemEventDb:         systemEventDb,
		validatorSeqDb:        validatorSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
	}
}

type sessionSystemEventCreatorTask struct {
	cfg *config.Config

	syncablesDb           store.Syncables
	systemEventDb         store.SystemEvents
	validatorSeqDb        store.ValidatorSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
}

func (t *sessionSystemEventCreatorTask) GetName() string {
	return TaskNameSessionSystemEventCreator
}

func (t *sessionSystemEventCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInSession {
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", "Analyzer", t.GetName(), payload.CurrentHeight))

	prevSeqs, err := t.getPrevValidatorSessionSequences(payload)
	if err != nil {
		return err
	}
	lastSessionHeight, err := t.getLastSessionHeight(payload)
	if err != nil {
		return err
	}

	activeSetPresenceChangeSystemEvents, err := t.getActiveSetPresenceChangeSystemEvents(payload.ValidatorSessionSequences, prevSeqs, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, activeSetPresenceChangeSystemEvents...)

	missedBlocksSystemEvents, err := t.getMissedBlocksSystemEvents(payload.ValidatorSessionSequences, lastSessionHeight, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, missedBlocksSystemEvents...)

	return nil
}

// NewEraSystemEventCreatorTask creates system events
func NewEraSystemEventCreatorTask(cfg *config.Config, accountEraSeqDb store.AccountEraSeq, validatorEraSeqDb store.ValidatorEraSeq) *eraSystemEventCreatorTask {
	return &eraSystemEventCreatorTask{
		cfg:               cfg,
		accountEraSeqDb:   accountEraSeqDb,
		validatorEraSeqDb: validatorEraSeqDb,
	}
}

type eraSystemEventCreatorTask struct {
	cfg *config.Config

	accountEraSeqDb   store.AccountEraSeq
	validatorEraSeqDb store.ValidatorEraSeq
}

func (t *eraSystemEventCreatorTask) GetName() string {
	return TaskNameEraSystemEventCreator
}

func (t *eraSystemEventCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	if !payload.Syncable.LastInEra {
		return nil
	}

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", "Analyzer", t.GetName(), payload.CurrentHeight))

	prevEraAccountSeqs, err := t.getPrevEraAccountSequences(payload)
	if err != nil {
		return err
	}
	delegationChangedSystemEvents, err := t.getDelegationChangedSystemEvents(payload.AccountEraSequences, prevEraAccountSeqs, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, delegationChangedSystemEvents...)

	prevEraValidatorSeqs, err := t.getPrevValidatorEraSequences(payload)
	if err != nil {
		return err
	}

	commissionChangeSystemEvents, err := t.getCommissionChangeSystemEvents(payload.ValidatorEraSequences, prevEraValidatorSeqs, payload.Syncable)
	if err != nil {
		return err
	}

	payload.SystemEvents = append(payload.SystemEvents, commissionChangeSystemEvents...)

	return nil
}

func (t *systemEventCreatorTask) getPrevHeightValidatorSequences(payload *payload) ([]model.ValidatorSeq, error) {
	var prevValidatorSeqs []model.ValidatorSeq

	if payload.CurrentHeight > t.cfg.FirstBlockHeight {
		var err error
		prevValidatorSeqs, err = t.validatorSeqDb.FindAllByHeight(payload.CurrentHeight - 1)

		if err != nil {
			return nil, err
		}
	}

	return prevValidatorSeqs, nil
}

func (t *eraSystemEventCreatorTask) getPrevEraAccountSequences(payload *payload) ([]model.AccountEraSeq, error) {
	var prevEraAccountSequences []model.AccountEraSeq

	if payload.CurrentHeight > t.cfg.FirstBlockHeight && payload.Syncable.Era > 1 {
		var err error
		prevEraAccountSequences, err = t.accountEraSeqDb.FindByEra(payload.Syncable.Era - 1)
		if err != nil {
			return nil, err
		}
	}

	return prevEraAccountSequences, nil
}

func (t *sessionSystemEventCreatorTask) getPrevValidatorSessionSequences(payload *payload) ([]model.ValidatorSessionSeq, error) {
	var prevValidatorSessionSequences []model.ValidatorSessionSeq

	if payload.CurrentHeight > t.cfg.FirstBlockHeight {
		var err error
		prevValidatorSessionSequences, err = t.validatorSessionSeqDb.FindBySession(payload.Syncable.Session - 1)
		if err != nil {
			return nil, err
		}
	}

	return prevValidatorSessionSequences, nil
}

func (t *sessionSystemEventCreatorTask) getLastSessionHeight(payload *payload) (int64, error) {
	lastSyncableInPrevSession, err := t.syncablesDb.FindLastInSession(payload.Syncable.Session - 1)
	var lastSessionHeight int64

	if err == store.ErrNotFound {
		lastSessionHeight = t.cfg.FirstBlockHeight
	} else if err != nil {
		return lastSessionHeight, err
	} else {
		lastSessionHeight = lastSyncableInPrevSession.Height
	}
	return lastSessionHeight, nil
}

func (t *sessionSystemEventCreatorTask) getMissedBlocksSystemEvents(currSeqs []model.ValidatorSessionSeq, lastSessionHeight int64, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	since := syncable.Session - missedConsecutiveThreshold
	if since < 0 {
		return systemEvents, nil
	}

	var missed int64
	for _, seq := range currSeqs {
		if seq.Online {
			continue
		}
		missed = 0
		kind := model.SystemEventMissedNConsecutive
		prevMissedEvents, err := t.systemEventDb.FindByActor(seq.StashAccount, &kind, &lastSessionHeight)
		if err != nil {
			return nil, err
		}
		if len(prevMissedEvents) > 0 {
			data := &model.MissedNConsecutive{}
			err := json.Unmarshal(prevMissedEvents[0].Data.RawMessage, data)
			if err != nil {
				return nil, err
			}
			missed = data.Missed
		}
		missed++

		if missed >= missedConsecutiveThreshold {
			newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventMissedNConsecutive, model.MissedNConsecutive{
				Missed:    missed,
				Threshold: missedConsecutiveThreshold,
			})
			if err != nil {
				return nil, err
			}

			systemEvents = append(systemEvents, newSystemEvent)
		}
	}
	return systemEvents, nil
}

func (t *sessionSystemEventCreatorTask) getActiveSetPresenceChangeSystemEvents(currSeqs, prevSeqs []model.ValidatorSessionSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	sort.Slice(currSeqs, func(i, j int) bool {
		return currSeqs[i].StashAccount < currSeqs[j].StashAccount
	})

	sort.Slice(prevSeqs, func(i, j int) bool {
		return prevSeqs[i].StashAccount < prevSeqs[j].StashAccount
	})

	var i, j int
	for i < len(currSeqs) || j < len(prevSeqs) {
		if (i >= len(currSeqs)) || (j < len(prevSeqs) && prevSeqs[j].StashAccount < currSeqs[i].StashAccount) {
			newSystemEvent, err := newSystemEvent(prevSeqs[j].StashAccount, syncable, model.SystemEventLeftSet, nil)
			if err != nil {
				return nil, err
			}
			systemEvents = append(systemEvents, newSystemEvent)
			j++
		} else if j >= len(prevSeqs) || prevSeqs[j].StashAccount > currSeqs[i].StashAccount {
			newSystemEvent, err := newSystemEvent(currSeqs[i].StashAccount, syncable, model.SystemEventJoinedSet, nil)
			if err != nil {
				return nil, err
			}
			systemEvents = append(systemEvents, newSystemEvent)
			i++
		} else {
			i++
			j++
		}
	}

	return systemEvents, nil
}

type delgationLookup map[string]struct{}

func (t *eraSystemEventCreatorTask) getDelegationChangedSystemEvents(currSeqs, prevSeqs []model.AccountEraSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	prevDelegationsforValidator := make(map[string]delgationLookup, len(prevSeqs))
	for _, seq := range prevSeqs {
		delegations, ok := prevDelegationsforValidator[seq.ValidatorStashAccount]
		if !ok {
			delegations = make(delgationLookup)
		}
		delegations[seq.StashAccount] = struct{}{}
		prevDelegationsforValidator[seq.ValidatorStashAccount] = delegations
	}

	currDelegationsforValidator := make(map[string]delgationLookup, len(currSeqs))
	for _, seq := range currSeqs {
		delegations, ok := currDelegationsforValidator[seq.ValidatorStashAccount]
		if !ok {
			delegations = make(delgationLookup)
		}
		delegations[seq.StashAccount] = struct{}{}
		currDelegationsforValidator[seq.ValidatorStashAccount] = delegations
	}

	var joined []string
	for v, currDelegations := range currDelegationsforValidator {
		prevDelegations, ok := prevDelegationsforValidator[v]
		if !ok {
			// validator wasnt active in previous era
			continue
		}

		joined = []string{}
		for d := range currDelegations {
			if _, ok = prevDelegations[d]; !ok {
				joined = append(joined, d)
			}
		}

		if len(joined) == 0 {
			continue
		}

		newSystemEvent, err := newSystemEvent(v, syncable, model.SystemEventDelegationJoined, model.DelegationChangeData{
			StashAccounts: joined,
		})
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	var left []string
	for v, prevDelegations := range prevDelegationsforValidator {
		currDelegations, ok := currDelegationsforValidator[v]
		if !ok {
			// validator wasnt active in current era
			continue
		}

		left = []string{}
		for d := range prevDelegations {
			if _, ok = currDelegations[d]; !ok {
				left = append(left, d)
			}
		}

		if len(left) == 0 {
			continue
		}

		newSystemEvent, err := newSystemEvent(v, syncable, model.SystemEventDelegationLeft, model.DelegationChangeData{
			StashAccounts: left,
		})
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	return systemEvents, nil
}

func (t *eraSystemEventCreatorTask) getPrevValidatorEraSequences(payload *payload) (prevEraSequences []model.ValidatorEraSeq, err error) {
	if payload.CurrentHeight > t.cfg.FirstBlockHeight {
		prevEraSequences, err = t.validatorEraSeqDb.FindByEra(payload.Syncable.Era - 1)
	}

	return prevEraSequences, err
}

func (t *systemEventCreatorTask) getActiveBalanceChangeSystemEvents(currValidatorSeqs, prevValidatorSeqs []model.ValidatorSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	prevHeightLookup := make(map[string]model.ValidatorSeq, len(prevValidatorSeqs))
	for _, seq := range prevValidatorSeqs {
		prevHeightLookup[seq.StashAccount] = seq
	}

	for _, validatorSequence := range currValidatorSeqs {
		if prevValidatorSequence, ok := prevHeightLookup[validatorSequence.StashAccount]; ok {
			newSystemEvent, err := t.getActiveBalanceChange(validatorSequence, prevValidatorSequence, syncable)
			if err != nil {
				if err != ErrActiveBalanceOutsideOfRange {
					return nil, err
				}
			} else {
				logger.Debug(fmt.Sprintf("active balance change for address %s occured [kind=%s]", validatorSequence.StashAccount, newSystemEvent.Kind))
				systemEvents = append(systemEvents, newSystemEvent)
			}
		}
	}
	return systemEvents, nil
}

func (t *systemEventCreatorTask) getActiveBalanceChange(currValidatorSeq model.ValidatorSeq, prevValidatorSeq model.ValidatorSeq, syncable *model.Syncable) (model.SystemEvent, error) {
	currValue := currValidatorSeq.ActiveBalance.Int64()
	prevValue := prevValidatorSeq.ActiveBalance.Int64()
	roundedChangeRate := getRoundedChangeRate(currValue, prevValue)
	roundedAbsChangeRate := math.Abs(roundedChangeRate)

	var kind model.SystemEventKind
	if roundedAbsChangeRate >= 0.1 && roundedAbsChangeRate < 1 {
		kind = model.SystemEventActiveBalanceChange1
	} else if roundedAbsChangeRate >= 1 && roundedAbsChangeRate < 10 {
		kind = model.SystemEventActiveBalanceChange2
	} else if roundedAbsChangeRate >= 10 {
		kind = model.SystemEventActiveBalanceChange3
	} else {
		return model.SystemEvent{}, ErrActiveBalanceOutsideOfRange
	}

	return newSystemEvent(currValidatorSeq.StashAccount, syncable, kind, model.PercentChangeData{
		Before: prevValue,
		After:  currValue,
		Change: roundedChangeRate,
	})
}

func (t *eraSystemEventCreatorTask) getCommissionChangeSystemEvents(currSeqs, prevSeqs []model.ValidatorEraSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	prevHeightLookup := make(map[string]model.ValidatorEraSeq, len(prevSeqs))
	for _, seq := range prevSeqs {
		prevHeightLookup[seq.StashAccount] = seq
	}

	for _, validatorSequence := range currSeqs {
		if prevValidatorSequence, ok := prevHeightLookup[validatorSequence.StashAccount]; ok {
			newSystemEvent, err := t.getCommissionChange(validatorSequence, prevValidatorSequence, syncable)
			if err != nil {
				if err != ErrCommissionOutsideOfRange {
					return nil, err
				}
			} else {
				logger.Debug(fmt.Sprintf("commission change for address %s occured [kind=%s]", validatorSequence.StashAccount, newSystemEvent.Kind))
				systemEvents = append(systemEvents, newSystemEvent)
			}
		}
	}
	return systemEvents, nil
}

func (t *eraSystemEventCreatorTask) getCommissionChange(currSeq, prevSeq model.ValidatorEraSeq, syncable *model.Syncable) (model.SystemEvent, error) {
	currValue := currSeq.Commission
	prevValue := prevSeq.Commission
	roundedChangeRate := getRoundedChangeRate(currValue, prevValue)
	roundedAbsChangeRate := math.Abs(roundedChangeRate)

	var kind model.SystemEventKind
	if roundedAbsChangeRate >= 0.1 && roundedAbsChangeRate < 1 {
		kind = model.SystemEventCommissionChange1
	} else if roundedAbsChangeRate >= 1 && roundedAbsChangeRate < 10 {
		kind = model.SystemEventCommissionChange2
	} else if roundedAbsChangeRate >= 10 {
		kind = model.SystemEventCommissionChange3
	} else {
		return model.SystemEvent{}, ErrCommissionOutsideOfRange
	}

	return newSystemEvent(currSeq.StashAccount, syncable, kind, model.PercentChangeData{
		Before: prevValue,
		After:  currValue,
		Change: roundedChangeRate,
	})
}

func getRoundedChangeRate(currValue int64, prevValue int64) float64 {
	var changeRate float64

	if prevValue == 0 {
		changeRate = float64(currValue)
	} else {
		changeRate = (float64(1) - (float64(currValue) / float64(prevValue))) * 100
	}

	roundedChangeRate := math.Round(changeRate/0.1) * 0.1
	return roundedChangeRate
}

func newSystemEvent(stashAccount string, syncable *model.Syncable, kind model.SystemEventKind, data interface{}) (model.SystemEvent, error) {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return model.SystemEvent{}, err
	}

	return model.SystemEvent{
		Height: syncable.Height,
		Time:   syncable.Time,
		Actor:  stashAccount,
		Kind:   kind,
		Data:   types.Jsonb{RawMessage: marshaledData},
	}, nil
}
