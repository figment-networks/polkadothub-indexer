package errors

import (
	"github.com/rollbar/rollbar-go"
	"github.com/figment-networks/polkadothub-indexer/config"
)

func RecoverError() {
	err := recover()
	rollbar.LogPanic(err, true)
}

func init() {
	rollbar.SetToken(config.RollbarAccessToken())
	rollbar.SetEnvironment(config.GoEnvironment())
	rollbar.SetServerRoot("github.com/figment-networks/polkadothub-indexer")
}
