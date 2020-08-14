package transaction

import (
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
)

type ListItem struct {
	Signature string `json:"public_key"`
	PublicKey string `json:"public_key"`
	Nonce     int64  `json:"nonce"`
	Method    string `json:"method"`
	Section   string `json:"method"`
	Args      string `json:"method"`
}

type ListView struct {
	Items []ListItem `json:"items"`
}

func ToListView(rawTransactions []*transactionpb.Transaction) *ListView {
	var items []ListItem
	for _, rawTransaction := range rawTransactions {
		item := ListItem{
			Signature: rawTransaction.GetSignature(),
			PublicKey: rawTransaction.GetSigner(),
			Nonce:     rawTransaction.GetNonce(),
			Method:    rawTransaction.GetMethod(),
			Section:   rawTransaction.GetSection(),
			Args:      rawTransaction.GetArgs(),
		}

		items = append(items, item)
	}

	return &ListView{
		Items: items,
	}
}