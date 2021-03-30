package transaction

import (
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
)

type TransactionListItem struct {
	Signature  string `json:"signature"`
	PublicKey  string `json:"public_key"`
	Nonce      int64  `json:"nonce"`
	Method     string `json:"method"`
	Section    string `json:"section"`
	Args       string `json:"args"`
	IsSuccess  bool   `json:"is_success"`
	PartialFee string `json:"partial_fee"`
	Tip        string `json:"tip"`
}

// swagger:response TransactionsView
type ListView struct {
	Items []TransactionListItem `json:"items"`
}

func ToListView(rawTransactions []*transactionpb.Transaction) *ListView {
	var items []TransactionListItem
	for _, rawTransaction := range rawTransactions {
		items = append(items, TransactionListItem{
			Signature:  rawTransaction.GetSignature(),
			PublicKey:  rawTransaction.GetSigner(),
			Nonce:      rawTransaction.GetNonce(),
			Method:     rawTransaction.GetMethod(),
			Section:    rawTransaction.GetSection(),
			Args:       rawTransaction.GetArgs(),
			IsSuccess:  rawTransaction.GetIsSuccess(),
			PartialFee: rawTransaction.GetPartialFee(),
			Tip:        rawTransaction.GetTip(),
		})
	}

	return &ListView{
		Items: items,
	}
}
