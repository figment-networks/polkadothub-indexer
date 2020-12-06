package model

import (
	"encoding/json"
	"errors"

	"github.com/figment-networks/polkadothub-indexer/types"
)

const (
	methodReward   = "Reward"
	sectionStaking = "staking"

	accountKey = "AccountId"
	balanceKey = "Balance"
)

var (
	errIncompatibleType     = errors.New("incompatible event type")
	errUnexpectedDataFormat = errors.New("unexpected event data format")
)

type EventSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	// Indexed data
	Index          int64       `json:"index"`
	ExtrinsicIndex int64       `json:"extrinsic_index"`
	Data           types.Jsonb `json:"data"`
	Phase          string      `json:"phase"`
	Method         string      `json:"method"`
	Section        string      `json:"section"`
}

type EventSeqWithTxHash struct {
	Height  int64       `json:"height"`
	Data    types.Jsonb `json:"data"`
	Method  string      `json:"method"`
	Section string      `json:"section"`
	TxHash  string      `json:"hash"`
}

func (EventSeq) TableName() string {
	return "event_sequences"
}

func (b *EventSeq) Valid() bool {
	return b.Sequence.Valid()
}

func (b *EventSeq) Equal(m EventSeq) bool {
	return b.Sequence.Equal(*m.Sequence)
}

func (b *EventSeq) IsReward() bool {
	return b.Method == methodReward && b.Section == sectionStaking
}

// EventData is format of Data for some event types (eg. rewards)
type EventData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (b *EventSeq) GetStashAndAmountFromData() (stash string, amount types.Quantity, err error) {
	if !b.IsReward() {
		return stash, amount, errIncompatibleType
	}

	var data []EventData
	err = json.Unmarshal(b.Data.RawMessage, &data)
	if err != nil {
		return stash, amount, err
	}

	for _, d := range data {
		if d.Name == accountKey {
			stash = d.Value
		} else if d.Name == balanceKey {
			amount, err = types.NewQuantityFromString(d.Value)
		}
	}
	if stash == "" {
		err = errUnexpectedDataFormat
	}
	return
}
