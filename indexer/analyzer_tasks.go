package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/pkg/errors"
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

	valueChangeSystemEvents, err := t.getValueChangeSystemEvents(payload.ValidatorSequences, prevHeightValidatorSeqs, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, valueChangeSystemEvents...)
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

	currActiveSeqs := append(payload.NewValidatorSessionSequences, payload.UpdatedValidatorSessionSequences...)
	prevSessionActiveSeqs, err := t.getPrevValidatorSessionSequences(payload)
	if err != nil {
		return err
	}
	lastSessionHeight, err := t.getLastSessionHeight(payload)
	if err != nil {
		return err
	}
	prevSessionSeqs, err := t.validatorSeqDb.FindAllByHeight(lastSessionHeight)
	if err != nil {
		return err
	}

	activeSetPresenceChangeSystemEvents, err := t.getActiveSetPresenceChangeSystemEvents(payload.ValidatorSequences, prevSessionSeqs, currActiveSeqs, prevSessionActiveSeqs, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, activeSetPresenceChangeSystemEvents...)

	missedBlocksSystemEvents, err := t.getMissedBlocksSystemEvents(currActiveSeqs, lastSessionHeight, payload.Syncable)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, missedBlocksSystemEvents...)

	return nil
}

// NewEraSystemEventCreatorTask creates system events
func NewEraSystemEventCreatorTask(cfg *config.Config, accountEraSeqDb store.AccountEraSeq) *eraSystemEventCreatorTask {
	return &eraSystemEventCreatorTask{
		cfg:             cfg,
		accountEraSeqDb: accountEraSeqDb,
	}
}

type eraSystemEventCreatorTask struct {
	cfg *config.Config

	accountEraSeqDb store.AccountEraSeq
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

func (t *sessionSystemEventCreatorTask) getActiveSetPresenceChangeSystemEvents(currSeqs, prevSeqs []model.ValidatorSeq, currActiveSeqs, prevActiveSeqs []model.ValidatorSessionSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent
	active := "a"
	waiting := "w"

	type status struct {
		was string
		is  string
	}

	lookup := make(map[string]*status)
	for _, seq := range prevActiveSeqs {
		lookup[seq.StashAccount] = &status{was: active}
	}
	for _, seq := range prevSeqs {
		if _, ok := lookup[seq.StashAccount]; !ok {
			lookup[seq.StashAccount] = &status{was: waiting}
		}
	}
	for _, seq := range currActiveSeqs {
		if s, ok := lookup[seq.StashAccount]; ok {
			s.is = active
		} else {
			lookup[seq.StashAccount] = &status{is: active}
			newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventJoinedActiveSet, nil)
			if err != nil {
				return nil, err
			}
			systemEvents = append(systemEvents, newSystemEvent)
		}
	}
	for _, seq := range currSeqs {
		if s, ok := lookup[seq.StashAccount]; ok {
			if s.is == "" {
				s.is = waiting
			}
			if s.is == s.was {
				continue
			} else if s.is == active {
				newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventJoinedActiveSet, nil)
				if err != nil {
					return nil, err
				}
				systemEvents = append(systemEvents, newSystemEvent)
			} else {
				newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventJoinedWaitingSet, nil)
				if err != nil {
					return nil, err
				}
				systemEvents = append(systemEvents, newSystemEvent)
			}
		} else {
			lookup[seq.StashAccount] = &status{is: waiting}
			newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventJoinedWaitingSet, nil)
			if err != nil {
				return nil, err
			}
			systemEvents = append(systemEvents, newSystemEvent)
		}
	}

	for _, seq := range prevSeqs {
		s, _ := lookup[seq.StashAccount]
		if s.is != "" {
			continue
		}
		newSystemEvent, err := newSystemEvent(seq.StashAccount, syncable, model.SystemEventLeftSet, nil)
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	return systemEvents, nil
}

type delgationLookup map[string]struct{}

func (t *eraSystemEventCreatorTask) getDelegationChangedSystemEvents(currSeqs, prevSeqs []model.AccountEraSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
	var systemEvents []model.SystemEvent

	lookupKey := func(account model.AccountEraSeq) string {
		return fmt.Sprintf("%v:%v", account.ValidatorStashAccount, account.StashAccount)
	}
	splitKey := func(key string) (string, string) {
		parts := strings.Split(key, ":")
		return parts[0], parts[1]
	}

	prevValidatorLookup := make(map[string]struct{})
	prevMcurr := make(map[string]struct{}, len(prevSeqs)) // set of prev minus current
	for _, seq := range prevSeqs {
		prevMcurr[lookupKey(seq)] = struct{}{}
		prevValidatorLookup[seq.ValidatorStashAccount] = struct{}{}
	}

	joinedDelegations := make(map[string][]string)
	currValidatorLookup := make(map[string]struct{})
	var v, d string
	for _, seq := range currSeqs {
		v = seq.ValidatorStashAccount
		d = seq.StashAccount

		currValidatorLookup[v] = struct{}{}
		if _, ok := prevValidatorLookup[v]; !ok {
			// validator not present in previous session, don't create delegation events
			continue
		}

		key := lookupKey(seq)
		if _, ok := prevMcurr[key]; ok {
			delete(prevMcurr, key)
			continue
		}

		joined, ok := joinedDelegations[v]
		if !ok {
			joined = []string{}
		}
		joinedDelegations[v] = append(joined, d)
	}

	leftDelegations := make(map[string][]string)
	for key := range prevMcurr {
		v, d = splitKey(key)
		if _, ok := currValidatorLookup[v]; !ok {
			// validator not present in current session, don't create delegation events
			continue
		}

		left, ok := leftDelegations[v]
		if !ok {
			left = []string{}
		}
		leftDelegations[v] = append(left, d)
	}

	for v, d := range joinedDelegations {
		newSystemEvent, err := newSystemEvent(v, syncable, model.SystemEventDelegationJoined, model.DelegationChangeData{
			StashAccounts: d,
		})
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	for v, d := range leftDelegations {
		newSystemEvent, err := newSystemEvent(v, syncable, model.SystemEventDelegationLeft, model.DelegationChangeData{
			StashAccounts: d,
		})
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	return systemEvents, nil
}

func (t *systemEventCreatorTask) getValueChangeSystemEvents(currValidatorSeqs, prevValidatorSeqs []model.ValidatorSeq, syncable *model.Syncable) ([]model.SystemEvent, error) {
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

			newSystemEvent, err = t.getCommissionChange(validatorSequence, prevValidatorSequence, syncable)
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

func (t *systemEventCreatorTask) getActiveBalanceChange(currValidatorSeq model.ValidatorSeq, prevValidatorSeq model.ValidatorSeq, syncable *model.Syncable) (model.SystemEvent, error) {
	currValue := currValidatorSeq.ActiveBalance.Int64()
	prevValue := prevValidatorSeq.ActiveBalance.Int64()
	roundedChangeRate := t.getRoundedChangeRate(currValue, prevValue)
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

func (t *systemEventCreatorTask) getCommissionChange(currValidatorSeq model.ValidatorSeq, prevValidatorSeq model.ValidatorSeq, syncable *model.Syncable) (model.SystemEvent, error) {
	currValue := currValidatorSeq.Commission.Int64()
	prevValue := prevValidatorSeq.Commission.Int64()
	roundedChangeRate := t.getRoundedChangeRate(currValue, prevValue)
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

	return newSystemEvent(currValidatorSeq.StashAccount, syncable, kind, model.PercentChangeData{
		Before: prevValue,
		After:  currValue,
		Change: roundedChangeRate,
	})
}

func (t *systemEventCreatorTask) getRoundedChangeRate(currValue int64, prevValue int64) float64 {
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
