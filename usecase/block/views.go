package block

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
)

type DetailsView struct {
	Chain          string                 `json:"chain"`
	SpecVersion    string                 `json:"spec_version"`
	Session        int64                  `json:"session"`
	Era            int64                  `json:"era"`
	Height         int64                  `json:"height"`
	Time           types.Time             `json:"time"`
	ParentHash     string                 `json:"parent_hash"`
	ExtrinsicsRoot string                 `json:"extrinsics_root"`
	StateRoot      string                 `json:"state_root"`
	Extrinsics     []ExtrinsicDetailsView `json:"extrinsics"`
}

type ExtrinsicDetailsView struct {
	ExtrinsicIndex int64  `json:"extrinsic_index"`
	IsSigned       bool   `json:"is_signed"`
	Signature      string `json:"signature"`
	PublicKey      string `json:"public_key"`
	Nonce          int64  `json:"nonce"`
	Method         string `json:"method"`
	Section        string `json:"section"`
	Args           string `json:"args"`
}

func ToDetailsView(rawResponse *blockpb.GetByHeightResponse) *DetailsView {
	rawBlock := rawResponse.GetBlock()

	view := &DetailsView{
		Chain:          rawResponse.GetChain(),
		SpecVersion:    rawResponse.GetSpecVersion(),
		Era:            rawResponse.GetEra(),
		Session:        rawResponse.GetSession(),
		Height:         rawBlock.GetHeader().GetHeight(),
		Time:           *types.NewTimeFromTimestamp(*rawBlock.GetHeader().GetTime()),
		ParentHash:     rawBlock.GetHeader().GetParentHash(),
		ExtrinsicsRoot: rawBlock.GetHeader().GetExtrinsicsRoot(),
		StateRoot:      rawBlock.GetHeader().GetStateRoot(),
	}

	for _, rawExtrinsic := range rawBlock.GetExtrinsics() {
		view.Extrinsics = append(view.Extrinsics, ExtrinsicDetailsView{
			ExtrinsicIndex: rawExtrinsic.GetExtrinsicIndex(),
			IsSigned:       rawExtrinsic.GetIsSignedTransaction(),
			Signature:      rawExtrinsic.GetSignature(),
			PublicKey:      rawExtrinsic.GetSigner(),
			Nonce:          rawExtrinsic.GetNonce(),
			Method:         rawExtrinsic.GetMethod(),
			Section:        rawExtrinsic.GetSection(),
			Args:           rawExtrinsic.GetArgs(),
		})
	}

	return view
}
