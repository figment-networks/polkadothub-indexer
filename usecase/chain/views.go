package chain

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/chain/chainpb"
)

type DetailsView struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
	GoVersion  string `json:"go_version"`

	ClientInfo string `json:"client_info"`

	ChainName string `json:"chain_name"`
	ChainType string `json:"chain_type"`

	NodeName         string            `json:"node_name"`
	NodeVersion      string            `json:"node_version"`
	NodeLocalPeerUID string            `json:"node_local_peer_uid"`
	NodeHealth       string            `json:"node_health"`
	NodeRoles        []string          `json:"node_roles"`
	NodeProperties   map[string]string `json:"node_properties"`

	GenesisHash string `json:"genesis_hash"`

	IndexingStarted    bool       `json:"indexing_started"`
	LastIndexVersion   int64      `json:"last_index_version"`
	LastIndexedHeight  int64      `json:"last_indexed_height"`
	LastIndexedSession int64      `json:"last_indexed_session"`
	LastIndexedEra     int64      `json:"last_indexed_era"`
	LastSpecVersion    string     `json:"chain_spec_version"`
	LastIndexedTime    types.Time `json:"last_indexed_time"`
	LastIndexedAt      types.Time `json:"last_indexed_at"`
	Lag                int64      `json:"indexing_lag"`
}

func ToDetailsView(recentSyncable *model.Syncable, headResponse *chainpb.GetHeadResponse, statusResponse *chainpb.GetStatusResponse) *DetailsView {
	view := &DetailsView{
		AppName:    config.AppName,
		AppVersion: config.AppVersion,
		GoVersion:  config.GoVersion,

		ClientInfo: statusResponse.GetClientInfo(),

		ChainName: statusResponse.GetChainName(),
		ChainType: statusResponse.GetChainType(),

		NodeName:         statusResponse.GetNodeName(),
		NodeVersion:      statusResponse.GetNodeVersion(),
		NodeLocalPeerUID: statusResponse.GetNodeLocalPeerUid(),
		NodeHealth:       statusResponse.GetNodeHealth(),
		NodeRoles:        statusResponse.GetNodeRoles(),
		NodeProperties:   statusResponse.GetNodeProperties(),

		GenesisHash: statusResponse.GetGenesisHash(),
	}

	view.IndexingStarted = recentSyncable != nil

	if view.IndexingStarted {
		view.LastIndexVersion = recentSyncable.IndexVersion
		view.LastIndexedHeight = recentSyncable.Height
		view.LastIndexedSession = recentSyncable.Session
		view.LastIndexedEra = recentSyncable.Era
		view.LastIndexedTime = recentSyncable.Time
		view.LastIndexedAt = recentSyncable.CreatedAt
		view.Lag = headResponse.Height - recentSyncable.Height
	}

	return view
}
