package account

import (
	"encoding/json"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
	"github.com/pkg/errors"
)

var (
	ErrCouldNotMarshalJSON   = errors.New("could not marshal data to JSON")
	ErrUnmarshalingEventData = errors.New("error when trying to unmarshal event data")
)

type HeightDetailsView struct {
	Nonce           int64  `json:"nonce"`
	ReferendumCount int64  `json:"referendum_count"`
	Free            string `json:"free"`
	Reserved        string `json:"reserved"`
	MiscFrozen      string `json:"misc_frozen"`
	FeeFrozen       string `json:"fee_frozen"`
}

func ToHeightDetailsView(rawAccount *accountpb.Account) *HeightDetailsView {
	return &HeightDetailsView{
		Nonce:           rawAccount.GetNonce(),
		ReferendumCount: rawAccount.GetReferendumCount(),
		Free:            rawAccount.GetFree(),
		Reserved:        rawAccount.GetReserved(),
		MiscFrozen:      rawAccount.GetMiscFrozen(),
		FeeFrozen:       rawAccount.GetFeeFrozen(),
	}
}

type DetailsView struct {
	Address string `json:"address"`

	*Identity

	Transfers []*BalanceTransfer `json:"transfers"`
	Deposits  []*BalanceDeposit  `json:"deposits"`
	Bonded    []*Bonded          `json:"bonded"`
	Unbonded  []*Unbonded        `json:"unbonded"`
	Withdrawn []*Withdrawn       `json:"withdrawn"`
}

func ToDetailsView(address string, rawAccountIdentity *accountpb.AccountIdentity, balanceTransferModels []model.EventSeq, balanceDepositModels []model.EventSeq, bondedModels []model.EventSeq, unbondedModels []model.EventSeq, withdrawnModels []model.EventSeq) (*DetailsView, error) {
	view := &DetailsView{
		Address: address,

		Identity: ToIdentity(rawAccountIdentity),
	}

	transfers, err := ToBalanceTransfers(address, balanceTransferModels)
	if err != nil {
		return nil, err
	}
	view.Transfers = transfers

	deposits, err := ToBalanceDeposits(balanceDepositModels)
	if err != nil {
		return nil, err
	}
	view.Deposits = deposits

	bondedList, err := ToBondedList(bondedModels)
	if err != nil {
		return nil, err
	}
	view.Bonded = bondedList

	unbondedList, err := ToUnbondedList(unbondedModels)
	if err != nil {
		return nil, err
	}
	view.Unbonded = unbondedList

	withdrawnList, err := ToWithdrawnList(withdrawnModels)
	if err != nil {
		return nil, err
	}
	view.Withdrawn = withdrawnList

	return view, nil
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
		DisplayName: rawAccountIdentity.GetDisplayName(),
		LegalName:   rawAccountIdentity.GetLegalName(),
		WebName:     rawAccountIdentity.GetWebName(),
		RiotName:    rawAccountIdentity.GetRiotName(),
		EmailName:   rawAccountIdentity.GetEmailName(),
		TwitterName: rawAccountIdentity.GetTwitterName(),
		Image:       rawAccountIdentity.GetImage(),
	}
}

type EventData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BalanceTransfer struct {
	Hash        string `json:"hash"`
	Amount      string `json:"amount"`
	Kind        string `json:"kind"`
	Participant string `json:"participant"`
}

func ToBalanceTransfers(forAddress string, balanceTransferEvents []model.EventSeq) ([]*BalanceTransfer, error) {
	var balanceTransfers []*BalanceTransfer
	for _, eventSeq := range balanceTransferEvents {
		eventData, err := unmarshalEventData(eventSeq)
		if err != nil {
			return nil, err
		}

		fromAddress := eventData[0]
		toAddress := eventData[1]
		amount := eventData[2]

		newBalanceTransfer := &BalanceTransfer{
			Amount: amount.Value,
		}

		if fromAddress.Value == forAddress {
			newBalanceTransfer.Kind = "out"
			newBalanceTransfer.Participant = toAddress.Value
		} else if toAddress.Value == forAddress {
			newBalanceTransfer.Kind = "in"
			newBalanceTransfer.Participant = fromAddress.Value
		}

		balanceTransfers = append(balanceTransfers, newBalanceTransfer)
	}

	return balanceTransfers, nil
}

type BalanceDeposit struct {
	Amount string `json:"amount"`
}

func ToBalanceDeposits(balanceDepositsEvents []model.EventSeq) ([]*BalanceDeposit, error) {
	var balanceDeposits []*BalanceDeposit
	for _, eventSeq := range balanceDepositsEvents {
		eventData, err := unmarshalEventData(eventSeq)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newBalanceDeposit := &BalanceDeposit{
			Amount: amount.Value,
		}

		balanceDeposits = append(balanceDeposits, newBalanceDeposit)
	}

	return balanceDeposits, nil
}

type Bonded struct {
	Amount   string `json:"amount"`
	Receiver string `json:"receiver"`
}

func ToBondedList(bondedEvents []model.EventSeq) ([]*Bonded, error) {
	var bondedList []*Bonded
	for _, eventSeq := range bondedEvents {
		eventData, err := unmarshalEventData(eventSeq)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newBonded := &Bonded{
			Amount: amount.Value,
		}

		bondedList = append(bondedList, newBonded)
	}

	return bondedList, nil
}

type Unbonded struct {
	Amount string `json:"amount"`
}

func ToUnbondedList(bondedEvents []model.EventSeq) ([]*Unbonded, error) {
	var unbondedList []*Unbonded
	for _, eventSeq := range bondedEvents {
		eventData, err := unmarshalEventData(eventSeq)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newUnbonded := &Unbonded{
			Amount: amount.Value,
		}

		unbondedList = append(unbondedList, newUnbonded)
	}

	return unbondedList, nil
}

type Withdrawn struct {
	Amount string `json:"amount"`
}

func ToWithdrawnList(bondedEvents []model.EventSeq) ([]*Withdrawn, error) {
	var withdrawnList []*Withdrawn
	for _, eventSeq := range bondedEvents {
		eventData, err := unmarshalEventData(eventSeq)
		if err != nil {
			return nil, err
		}

		amount := eventData[1]

		newWithdrawn := &Withdrawn{
			Amount: amount.Value,
		}

		withdrawnList = append(withdrawnList, newWithdrawn)
	}

	return withdrawnList, nil
}

func unmarshalEventData(eventSeq model.EventSeq) ([]*EventData, error) {
	bytes, err := eventSeq.Data.RawMessage.MarshalJSON()
	if err != nil {
		err := ErrCouldNotMarshalJSON
		return nil, err
	}
	var eventData []*EventData
	if err := json.Unmarshal(bytes, &eventData); err != nil {
		return nil, ErrUnmarshalingEventData
	}
	return eventData, nil
}
