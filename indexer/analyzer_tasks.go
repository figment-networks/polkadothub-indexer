package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
	TaskNameSystemEventCreator = "SystemEventCreator"
)

var (
	ErrActiveBalanceOutsideOfRange = errors.New("active balance is outside of specified buckets")
	ErrCommissionOutsideOfRange    = errors.New("commission is outside of specified buckets")
)

// NewSystemEventCreatorTask creates system events
func NewSystemEventCreatorTask(cfg *config.Config, syncablesDb store.Syncables, validatorSeqDb store.ValidatorSeq, validatorSessionSeqDb store.ValidatorSessionSeq) *systemEventCreatorTask {
	return &systemEventCreatorTask{
		cfg:                   cfg,
		syncablesDb:           syncablesDb,
		validatorSeqDb:        validatorSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
	}
}

type systemEventCreatorTask struct {
	cfg                   *config.Config
	syncablesDb           store.Syncables
	validatorSeqDb        store.ValidatorSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
}

type systemEventRawData map[string]interface{}

func (t *systemEventCreatorTask) GetName() string {
	return TaskNameSystemEventCreator
}

func (t *systemEventCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", "Analyzer", t.GetName(), payload.CurrentHeight))

	currValidatorSeqs := append(payload.NewValidatorSequences, payload.UpdatedValidatorSequences...)
	prevValidatorSeqs, err := t.getPrevHeightValidatorSequences(payload)
	if err != nil {
		return err
	}

	valueChangeSystemEvents, err := t.getValueChangeSystemEvents(currValidatorSeqs, prevValidatorSeqs)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, valueChangeSystemEvents...)

	if !payload.Syncable.LastInSession {
		return nil
	}

	currActiveSeqs := append(payload.NewValidatorSessionSequences, payload.UpdatedValidatorSessionSequences...)
	prevActiveSeqs, err := t.getPrevValidatorSessionSequences(payload)
	if err != nil {
		return err
	}
	prevSessionSeqs, err := t.getPrevSessionValidatorSequences(payload)
	if err != nil {
		return err
	}

	activeSetPresenceChangeSystemEvents, err := t.getActiveSetPresenceChangeSystemEvents(currValidatorSeqs, prevSessionSeqs, currActiveSeqs, prevActiveSeqs, payload.Syncable.Height)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, activeSetPresenceChangeSystemEvents...)

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

func (t *systemEventCreatorTask) getPrevValidatorSessionSequences(payload *payload) ([]model.ValidatorSessionSeq, error) {
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

func (t *systemEventCreatorTask) getPrevSessionValidatorSequences(payload *payload) ([]model.ValidatorSeq, error) {
	var prevEraValidatorSequences []model.ValidatorSeq
	lastSyncableInPrevSession, err := t.syncablesDb.FindLastInSession(payload.Syncable.Session - 1)
	var lastSessionHeight int64
	// todo improve logic
	if err != nil {
		if err == store.ErrNotFound {
			lastSessionHeight = t.cfg.FirstBlockHeight
		} else {
			return nil, err
		}
	} else {
		lastSessionHeight = lastSyncableInPrevSession.Height
	}

	if payload.CurrentHeight > t.cfg.FirstBlockHeight {
		var err error
		prevEraValidatorSequences, err = t.validatorSeqDb.FindAllByHeight(lastSessionHeight)
		if err != nil {
			return nil, err
		}
	}
	return prevEraValidatorSequences, nil
}

func (t *systemEventCreatorTask) getActiveSetPresenceChangeSystemEvents(currSeqs, prevSeqs []model.ValidatorSeq, currActiveSeqs, prevActiveSeqs []model.ValidatorSessionSeq, currHeight int64) ([]*model.SystemEvent, error) {
	var systemEvents []*model.SystemEvent
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
			newSystemEvent, err := t.newSystemEvent(seq.StashAccount, currHeight, model.SystemEventJoinedActiveSet, systemEventRawData{})
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
				newSystemEvent, err := t.newSystemEvent(seq.StashAccount, currHeight, model.SystemEventJoinedActiveSet, systemEventRawData{})
				if err != nil {
					return nil, err
				}
				systemEvents = append(systemEvents, newSystemEvent)
			} else {
				newSystemEvent, err := t.newSystemEvent(seq.StashAccount, currHeight, model.SystemEventJoinedWaitingSet, systemEventRawData{})
				if err != nil {
					return nil, err
				}
				systemEvents = append(systemEvents, newSystemEvent)
			}
		} else {
			lookup[seq.StashAccount] = &status{is: waiting}
			newSystemEvent, err := t.newSystemEvent(seq.StashAccount, currHeight, model.SystemEventJoinedWaitingSet, systemEventRawData{})
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
		newSystemEvent, err := t.newSystemEvent(seq.StashAccount, currHeight, model.SystemEventLeftSet, systemEventRawData{})
		if err != nil {
			return nil, err
		}
		systemEvents = append(systemEvents, newSystemEvent)
	}

	return systemEvents, nil
}

func (t *systemEventCreatorTask) getValueChangeSystemEvents(currValidatorSeqs, prevValidatorSeqs []model.ValidatorSeq) ([]*model.SystemEvent, error) {
	var systemEvents []*model.SystemEvent

	prevHeightLookup := make(map[string]model.ValidatorSeq, len(prevValidatorSeqs))
	for _, seq := range prevValidatorSeqs {
		prevHeightLookup[seq.StashAccount] = seq
	}

	for _, validatorSequence := range currValidatorSeqs {
		if prevValidatorSequence, ok := prevHeightLookup[validatorSequence.StashAccount]; ok {
			newSystemEvent, err := t.getActiveBalanceChange(validatorSequence, prevValidatorSequence)
			if err != nil {
				if err != ErrActiveBalanceOutsideOfRange {
					return nil, err
				}
			} else {
				logger.Debug(fmt.Sprintf("active balance change for address %s occured [kind=%s]", validatorSequence.StashAccount, newSystemEvent.Kind))
				systemEvents = append(systemEvents, newSystemEvent)
			}

			newSystemEvent, err = t.getCommissionChange(validatorSequence, prevValidatorSequence)
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

func (t *systemEventCreatorTask) getActiveBalanceChange(currValidatorSeq model.ValidatorSeq, prevValidatorSeq model.ValidatorSeq) (*model.SystemEvent, error) {
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
		return nil, ErrActiveBalanceOutsideOfRange
	}

	return t.newSystemEvent(currValidatorSeq.StashAccount, currValidatorSeq.Height, kind, systemEventRawData{
		"before": prevValue,
		"after":  currValue,
		"change": roundedChangeRate,
	})
}

func (t *systemEventCreatorTask) getCommissionChange(currValidatorSeq model.ValidatorSeq, prevValidatorSeq model.ValidatorSeq) (*model.SystemEvent, error) {
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
		return nil, ErrCommissionOutsideOfRange
	}

	return t.newSystemEvent(currValidatorSeq.StashAccount, currValidatorSeq.Height, kind, systemEventRawData{
		"before": prevValue,
		"after":  currValue,
		"change": roundedChangeRate,
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

func (t *systemEventCreatorTask) newSystemEvent(stashAccount string, height int64, kind model.SystemEventKind, data map[string]interface{}) (*model.SystemEvent, error) {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &model.SystemEvent{
		Height: height,
		Actor:  stashAccount,
		Kind:   kind,
		Data:   types.Jsonb{RawMessage: marshaledData},
	}, nil
}
