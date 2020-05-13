package shared

import (
	"github.com/figment-networks/polkadothub-indexer/clients/grpcclient"
	"github.com/figment-networks/polkadothub-indexer/clients/timescaleclient"
	"github.com/figment-networks/polkadothub-indexer/config"
)

func NewProxyClient() *grpcclient.Client {
	return grpcclient.New(grpcclient.Config{Url: config.ProxyUrl()})
}

func NewDbClient() timescaleclient.DbClient {
	return timescaleclient.New(timescaleclient.Config{
		Host: config.DbHost(),
		User: config.DbUser(),
		Password: config.DbPassword(),
		DatabaseName: config.DbName(),
		SLLMode:      config.DbSSLMode(),
	})
}

