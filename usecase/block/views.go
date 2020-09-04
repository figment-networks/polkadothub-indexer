package block

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
)

type DetailsView struct {
	Height         int64                  `json:"height"`
	Time           types.Time             `json:"time"`
	Hash           string                 `json:"hash"`
	ParentHash     string                 `json:"parent_hash"`
	ExtrinsicsRoot string                 `json:"extrinsics_root"`
	StateRoot      string                 `json:"state_root"`
	Extrinsics     []ExtrinsicDetailsView `json:"extrinsics"`
}

type ExtrinsicDetailsView struct {
	ExtrinsicIndex int64  `json:"extrinsic_index"`
	Hash           string `json:"hash"`
	IsSigned       bool   `json:"is_signed"`
	Signature      string `json:"signature"`
	PublicKey      string `json:"public_key"`
	Nonce          int64  `json:"nonce"`
	Method         string `json:"method"`
	Section        string `json:"section"`
	Args           string `json:"args"`
	IsSuccess      bool   `json:"is_success"`
	PartialFee     string `json:"partial_fee"`
	Tip            string `json:"tip"`
}

func ToDetailsView(rawResponse *blockpb.GetByHeightResponse) *DetailsView {
	rawBlock := rawResponse.GetBlock()

	view := &DetailsView{
		Height:         rawBlock.GetHeader().GetHeight(),
		Time:           *types.NewTimeFromTimestamp(*rawBlock.GetHeader().GetTime()),
		Hash:           rawBlock.GetBlockHash(),
		ParentHash:     rawBlock.GetHeader().GetParentHash(),
		ExtrinsicsRoot: rawBlock.GetHeader().GetExtrinsicsRoot(),
		StateRoot:      rawBlock.GetHeader().GetStateRoot(),
	}

	for _, rawExtrinsic := range rawBlock.GetExtrinsics() {
		view.Extrinsics = append(view.Extrinsics, ExtrinsicDetailsView{
			ExtrinsicIndex: rawExtrinsic.GetExtrinsicIndex(),
			Hash:           rawExtrinsic.GetHash(),
			IsSigned:       rawExtrinsic.GetIsSignedTransaction(),
			Signature:      rawExtrinsic.GetSignature(),
			PublicKey:      rawExtrinsic.GetSigner(),
			Nonce:          rawExtrinsic.GetNonce(),
			Method:         rawExtrinsic.GetMethod(),
			Section:        rawExtrinsic.GetSection(),
			Args:           rawExtrinsic.GetArgs(),
			IsSuccess:      rawExtrinsic.GetIsSuccess(),
			PartialFee:     rawExtrinsic.GetPartialFee(),
			Tip:            rawExtrinsic.GetTip(),
		})
	}

	return view
}
