package model

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/figment-networks/polkadothub-indexer/types"
)

const (
	TxMethodPayoutStakers = "payoutStakers"
	TxSectionStaking      = "staking"
)

var (
	errIncompatibleTxArgs     = errors.New("incompatible transaction args")
	errUnexpectedTxDataFormat = errors.New("unexpected data format for transaction")
)

type TransactionSeq struct {
	ID types.ID `json:"id"`

	*Sequence

	// Indexed data
	Index   int64  `json:"index"`
	Hash    string `json:"hash"`
	Method  string `json:"method"`
	Section string `json:"section"`
	Args    string `json:"args"`
}

func (TransactionSeq) TableName() string {
	return "transaction_sequences"
}

func (t *TransactionSeq) Valid() bool {
	if t.Hash == "" || t.Method == "" || t.Section == "" || !t.Sequence.Valid() {
		return false
	}
	return true
}

func (t *TransactionSeq) IsPayoutStakers() bool {
	return t.Method == TxMethodPayoutStakers && t.Section == TxSectionStaking
}

func (t *TransactionSeq) GetStashAndEraFromPayoutArgs() (validatorStash string, era int64, err error) {
	if !t.IsPayoutStakers() {
		return validatorStash, era, errIncompatibleTxArgs
	}

	var data []string

	err = json.Unmarshal([]byte(t.Args), &data)
	if err != nil {
		return validatorStash, era, err
	}

	if len(data) < 2 {
		return validatorStash, era, errUnexpectedTxDataFormat
	}
	validatorStash = data[0]
	era, err = strconv.ParseInt(data[1], 10, 64)
	return
}
