package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/common"
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
)

var (
	ErrCouldNotMarshalJSON   = errors.New("could not marshal data to JSON")
	ErrUnmarshalingEventData = errors.New("error when trying to unmarshal event data")
)

const (
	balanceKey string = "Balance"
)

// swagger:response AccountHeightDetailsView
type HeightDetailsView struct {
	Nonce int64 `json:"nonce"`
	// Free balance of account
	Free string `json:"free"`
	// Reserved balance of account
	Reserved string `json:"reserved"`
	// MiscFrozen balance of account
	MiscFrozen string `json:"misc_frozen"`
	// FeeFrozen balance of account
	FeeFrozen string `json:"fee_frozen"`
}

func ToHeightDetailsView(rawAccount *accountpb.Account) *HeightDetailsView {
	return &HeightDetailsView{
		Nonce:      rawAccount.GetNonce(),
		Free:       rawAccount.GetFree(),
		Reserved:   rawAccount.GetReserved(),
		MiscFrozen: rawAccount.GetMiscFrozen(),
		FeeFrozen:  rawAccount.GetFeeFrozen(),
	}
}

// swagger:response AccountDetailsView
type DetailsView struct {
	// Address of account
	Address string `json:"address"`
	// Account is balance information for an account
	Account *Account `json:"account"`
	// Identity is identity details for an account
	*Identity

	// Transfers is a list of all balances.Transfer events for account
	Transfers []*BalanceTransfer `json:"transfers"`
	// Deposits is a list of all balances.Deposit events for account
	Deposits []*BalanceDeposit `json:"deposits"`
	// Bonded is a list of all staking.Bonded events for account
	Bonded []*Bonded `json:"bonded"`
	// Unbonded is a list of all staking.Unbonded events for account
	Unbonded []*Unbonded `json:"unbonded"`
	// Withdrawn is a list of all staking.Withdrawn events for account
	Withdrawn []*Withdrawn `json:"withdrawn"`
	// Delegations is a list of latest delegations for account
	Delegations []*common.Delegation `json:"delegations"`
}

func ToDetailsView(address string, rawAccountIdentity *accountpb.AccountIdentity, rawAccount *accountpb.Account, accountEraSeqs []model.AccountEraSeq, balanceTransferModels, balanceDepositModels, bondedModels, unbondedModels, withdrawnModels []model.EventSeqWithTxHash) (DetailsView, error) {
	view := DetailsView{
		Address:     address,
		Account:     ToAccount(rawAccount),
		Identity:    ToIdentity(rawAccountIdentity),
		Delegations: common.ToDelegations(accountEraSeqs),
	}

	transfers, err := ToBalanceTransfers(address, balanceTransferModels)
	if err != nil {
		return DetailsView{}, err
	}
	view.Transfers = transfers

	deposits, err := ToBalanceDeposits(balanceDepositModels)
	if err != nil {
		return DetailsView{}, err
	}
	view.Deposits = deposits

	bondedList, err := ToBondedList(bondedModels)
	if err != nil {
		return DetailsView{}, err
	}
	view.Bonded = bondedList

	unbondedList, err := ToUnbondedList(unbondedModels)
	if err != nil {
		return DetailsView{}, err
	}
	view.Unbonded = unbondedList

	withdrawnList, err := ToWithdrawnList(withdrawnModels)
	if err != nil {
		return DetailsView{}, err
	}
	view.Withdrawn = withdrawnList

	return view, nil
}

// swagger:response AccountRewardsView
type RewardsView struct {
	// Account is stash account
	Account string `json:"account"`
	// Start is start time of query period
	Start time.Time `json:"start_time"`
	// End is end time of query period
	End time.Time `json:"end_time"`
	// TotalAmount is summed total of rewards earned
	TotalAmount string `json:"total_amount"`
	// Rewards is list of rewards
	Rewards []Reward `json:"rewards"`
}

type Reward struct {
	// Height of block where reward event occured
	Height int64 `json:"height"`
	// Time of block where reward event occured
	Time time.Time `json:"time"`
	// Amount is reward balance claimed by account
	Amount string `json:"amount"`
}

func toRewardsView(events []model.EventSeq, account string, start, end time.Time) (RewardsView, error) {
	rewards := make([]Reward, len(events))
	totalAmount := types.NewQuantityFromInt64(0)
	for i, e := range events {
		eventData, err := unmarshalEventData(e.Data)
		if err != nil {
			return RewardsView{}, err
		}

		var amountStr string
		for _, val := range eventData {
			if val.Name == balanceKey {
				amountStr = val.Value
			}
		}

		if amountStr == "" {
			return RewardsView{}, fmt.Errorf("expected to find key %v in event data", balanceKey)
		}

		amount, err := types.NewQuantityFromString(amountStr)
		if err != nil {
			return RewardsView{}, fmt.Errorf("could not convert %v to amount", amountStr)
		}

		totalAmount.Add(amount)

		rewards[i] = Reward{
			Amount: amountStr,
			Height: e.Height,
			Time:   e.Time.UTC(),
		}
	}

	return RewardsView{
		Account:     account,
		Start:       start.UTC(),
		End:         end.UTC(),
		TotalAmount: totalAmount.String(),
		Rewards:     rewards,
	}, nil
}

type Identity struct {
	Deposit     string `json:"deposit"`
	DisplayName string `json:"display_name"`
	LegalName   string `json:"legal_name"`
	WebName     string `json:"web_name"`
	RiotName    string `json:"riot_name"`
	EmailName   string `json:"email_name"`
	TwitterName string `json:"twitter_name"`
	Image       string `json:"image"`
}

func ToIdentity(rawAccountIdentity *accountpb.AccountIdentity) *Identity {
	return &Identity{
		Deposit:     rawAccountIdentity.GetDeposit(),
		DisplayName: strings.TrimSpace(rawAccountIdentity.GetDisplayName()),
		LegalName:   strings.TrimSpace(rawAccountIdentity.GetLegalName()),
		WebName:     strings.TrimSpace(rawAccountIdentity.GetWebName()),
		RiotName:    strings.TrimSpace(rawAccountIdentity.GetRiotName()),
		EmailName:   strings.TrimSpace(rawAccountIdentity.GetEmailName()),
		TwitterName: strings.TrimSpace(rawAccountIdentity.GetTwitterName()),
		Image:       rawAccountIdentity.GetImage(),
	}
}

// Account is balance information for an account
type Account struct {
	// Free balance of account
	Free string `json:"free"`
	// Reserved balance of account
	Reserved string `json:"reserved"`
	// MiscFrozen balance of account
	MiscFrozen string `json:"misc_frozen"`
	// FeeFrozen balance of account
	FeeFrozen string `json:"fee_frozen"`
}

func ToAccount(rawAccount *accountpb.Account) *Account {
	return &Account{
		Free:       rawAccount.GetFree(),
		Reserved:   rawAccount.GetReserved(),
		MiscFrozen: rawAccount.GetMiscFrozen(),
		FeeFrozen:  rawAccount.GetFeeFrozen(),
	}
}

type EventData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BalanceTransfer struct {
	// Hash is transaction hash
	Hash string `json:"transaction_hash"`
	// Height is block height transaction occured
	Height int64 `json:"height"`
	// Amount is balance that was transferred
	Amount string `json:"amount"`
	// Kind is transfer kind (either "in" for transfer into account or "out" for transfer out of account)
	// example: in
	Kind string `json:"kind"`
	// Participant is account
	Participant string `json:"participant"`
}

func ToBalanceTransfers(forAddress string, balanceTransferEvents []model.EventSeqWithTxHash) ([]*BalanceTransfer, error) {
	balanceTransfers := make([]*BalanceTransfer, len(balanceTransferEvents))
	for i, eventSeq := range balanceTransferEvents {
		eventData, err := unmarshalEventData(eventSeq.Data)
		if err != nil {
			return nil, err
		}

		fromAddress := eventData[0]
		toAddress := eventData[1]
		amount := eventData[2]

		newBalanceTransfer := &BalanceTransfer{
			Amount: amount.Value,
			Height: eventSeq.Height,
			Hash:   eventSeq.TxHash,
		}

		if fromAddress.Value == forAddress {
			newBalanceTransfer.Kind = "out"
			newBalanceTransfer.Participant = toAddress.Value
		} else if toAddress.Value == forAddress {
			newBalanceTransfer.Kind = "in"
			newBalanceTransfer.Participant = fromAddress.Value
		}

		balanceTransfers[i] = newBalanceTransfer
	}

	return balanceTransfers, nil
}

type BalanceDeposit struct {
	// Amount is balance that was deposited into account
	Amount string `json:"amount"`
	// Hash is transaction hash
	Hash string `json:"transaction_hash"`
	// Height is block height transaction occured
	Height int64 `json:"height"`
}

func ToBalanceDeposits(balanceDepositsEvents []model.EventSeqWithTxHash) ([]*BalanceDeposit, error) {
	balanceDeposits := make([]*BalanceDeposit, len(balanceDepositsEvents))
	for i, eventSeq := range balanceDepositsEvents {
		eventData, err := unmarshalEventData(eventSeq.Data)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newBalanceDeposit := &BalanceDeposit{
			Amount: amount.Value,
			Hash:   eventSeq.TxHash,
			Height: eventSeq.Height,
		}

		balanceDeposits[i] = newBalanceDeposit
	}

	return balanceDeposits, nil
}

type Bonded struct {
	// Amount is balance that was bonded
	Amount string `json:"amount"`
	// Receiver is account that bonded
	Receiver string `json:"receiver"`
	// Hash is transaction hash
	Hash string `json:"transaction_hash"`
	// Height is block height transaction occured
	Height int64 `json:"height"`
}

func ToBondedList(bondedEvents []model.EventSeqWithTxHash) ([]*Bonded, error) {
	bondedList := make([]*Bonded, len(bondedEvents))
	for i, eventSeq := range bondedEvents {
		eventData, err := unmarshalEventData(eventSeq.Data)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newBonded := &Bonded{
			Amount:   amount.Value,
			Hash:     eventSeq.TxHash,
			Height:   eventSeq.Height,
			Receiver: eventData[0].Value,
		}

		bondedList[i] = newBonded
	}

	return bondedList, nil
}

type Unbonded struct {
	// Amount is balance that was unbonded
	Amount string `json:"amount"`
	// Hash is transaction hash
	Hash string `json:"transaction_hash"`
	// Height is block height transaction occured
	Height int64 `json:"height"`
}

func ToUnbondedList(bondedEvents []model.EventSeqWithTxHash) ([]*Unbonded, error) {
	unbondedList := make([]*Unbonded, len(bondedEvents))
	for i, eventSeq := range bondedEvents {
		eventData, err := unmarshalEventData(eventSeq.Data)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newUnbonded := &Unbonded{
			Amount: amount.Value,
			Hash:   eventSeq.TxHash,
			Height: eventSeq.Height,
		}

		unbondedList[i] = newUnbonded
	}

	return unbondedList, nil
}

type Withdrawn struct {
	// Amount is balance that was withdrawn
	Amount string `json:"amount"`
	// Hash is transaction hash
	Hash string `json:"transaction_hash"`
	// Height is block height transaction occured
	Height int64 `json:"height"`
}

func ToWithdrawnList(withdrawnEvents []model.EventSeqWithTxHash) ([]*Withdrawn, error) {
	withdrawnList := make([]*Withdrawn, len(withdrawnEvents))
	for i, eventSeq := range withdrawnEvents {
		eventData, err := unmarshalEventData(eventSeq.Data)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newWithdrawn := &Withdrawn{
			Amount: amount.Value,
			Hash:   eventSeq.TxHash,
			Height: eventSeq.Height,
		}

		withdrawnList[i] = newWithdrawn
	}

	return withdrawnList, nil
}

func unmarshalEventData(data types.Jsonb) ([]*EventData, error) {
	bytes, err := data.RawMessage.MarshalJSON()
	if err != nil {
		return nil, ErrCouldNotMarshalJSON
	}
	var eventData []*EventData
	if err := json.Unmarshal(bytes, &eventData); err != nil {
		return nil, ErrUnmarshalingEventData
	}
	return eventData, nil
}
