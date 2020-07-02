package account

import (
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
)

type DetailsView struct {
	Nonce           int64  `json:"nonce"`
	ReferendumCount int64  `json:"referendum_count"`
	Free            string `json:"free"`
	Reserved        string `json:"reserved"`
	MiscFrozen      string `json:"misc_frozen"`
	FeeFrozen       string `json:"fee_frozen"`
}

func ToDetailsView(rawAccount *accountpb.Account) *DetailsView {
	return &DetailsView{
		Nonce:           rawAccount.GetNonce(),
		ReferendumCount: rawAccount.GetReferendumCount(),
		Free:            rawAccount.GetFree(),
		Reserved:        rawAccount.GetReserved(),
		MiscFrozen:      rawAccount.GetMiscFrozen(),
		FeeFrozen:       rawAccount.GetFeeFrozen(),
	}
}

type IdentityDetailsView struct {
	Deposit     string `json:"deposit"`
	DisplayName string `json:"display_name"`
	LegalName   string `json:"legal_name"`
	WebName     string `json:"web_name"`
	RiotName    string `json:"riot_name"`
	EmailName   string `json:"email_name"`
	TwitterName string `json:"twitter_name"`
	Image       string `json:"image"`
}

func ToIdentityDetailsView(rawAccountIdentity *accountpb.AccountIdentity) *IdentityDetailsView {
	return &IdentityDetailsView{
		Deposit: rawAccountIdentity.GetDeposit(),
		DisplayName: rawAccountIdentity.GetDisplayName(),
		LegalName: rawAccountIdentity.GetLegalName(),
		WebName: rawAccountIdentity.GetWebName(),
		RiotName: rawAccountIdentity.GetRiotName(),
		EmailName: rawAccountIdentity.GetEmailName(),
		TwitterName: rawAccountIdentity.GetTwitterName(),
		Image: rawAccountIdentity.GetImage(),
	}
}
