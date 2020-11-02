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
)

// NewSystemEventCreatorTask creates system events
func NewSystemEventCreatorTask(cfg *config.Config, validatorSeqDb store.ValidatorSeq) *systemEventCreatorTask {
	return &systemEventCreatorTask{
		cfg:            cfg,
		validatorSeqDb: validatorSeqDb,
	}
}

type systemEventCreatorTask struct {
	cfg            *config.Config
	validatorSeqDb store.ValidatorSeq
}

type systemEventRawData map[string]interface{}

func (t *systemEventCreatorTask) GetName() string {
	return TaskNameSystemEventCreator
}

func (t *systemEventCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", "Analyzer", t.GetName(), payload.CurrentHeight))

	currHeightValidatorSequences := append(payload.NewValidatorSequences, payload.UpdatedValidatorSequences...)
	prevHeightValidatorSequences, err := t.getPrevHeightValidatorSequences(payload)
	if err != nil {
		return err
	}

	// balance changes
	valueChangeSystemEvents, err := t.getValueChangeSystemEvents(currHeightValidatorSequences, prevHeightValidatorSequences)
	if err != nil {
		return err
	}
	payload.SystemEvents = append(payload.SystemEvents, valueChangeSystemEvents...)

	return nil
}

func (t *systemEventCreatorTask) getPrevHeightValidatorSequences(payload *payload) ([]model.ValidatorSeq, error) {
	var prevHeightValidatorSequences []model.ValidatorSeq
	if payload.CurrentHeight > t.cfg.FirstBlockHeight {
		var err error
		prevHeightValidatorSequences, err = t.validatorSeqDb.FindAllByHeight(payload.CurrentHeight - 1)
		if err != nil {
			return nil, err
		}
	}
	return prevHeightValidatorSequences, nil
}

func (t *systemEventCreatorTask) getValueChangeSystemEvents(currHeightValidatorSequences, prevHeightValidatorSequences []model.ValidatorSeq) ([]*model.SystemEvent, error) {
	var systemEvents []*model.SystemEvent

	prevLookup := make(map[string]model.ValidatorSeq, len(prevHeightValidatorSequences))
	for _, seq := range prevHeightValidatorSequences {
		prevLookup[seq.StashAccount] = seq
	}

	for _, validatorSequence := range currHeightValidatorSequences {
		if prevValidatorSequence, ok := prevLookup[validatorSequence.StashAccount]; ok {
			newSystemEvent, err := t.getActiveBalanceChange(validatorSequence, prevValidatorSequence)
			if err != nil {
				if err != ErrActiveBalanceOutsideOfRange {
					return nil, err
				}
				continue
			}

			logger.Debug(fmt.Sprintf("active balance change for address %s occured [kind=%s]", validatorSequence.StashAccount, newSystemEvent.Kind))
			systemEvents = append(systemEvents, newSystemEvent)
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

	return t.newSystemEvent(currValidatorSeq, kind, systemEventRawData{
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

func (t *systemEventCreatorTask) newSystemEvent(seq model.ValidatorSeq, kind model.SystemEventKind, data map[string]interface{}) (*model.SystemEvent, error) {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &model.SystemEvent{
		Height: seq.Height,
		Actor:  seq.StashAccount,
		Kind:   kind,
		Data:   types.Jsonb{RawMessage: marshaledData},
	}, nil
}
