package server

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/apps/shared"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/usecases/ping"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-indexer/utils/log"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
)

func main() {
	defer errors.RecoverError()

	// CLIENTS
	node := shared.NewNodeClient()
	db := shared.NewDbClient()

	// HANDLERS
	pingHandler := ping.NewHttpHandler()

	router = gin.Default()
	router.GET("/ping", pingHandler.Handle)

	log.Info(fmt.Sprintf("Starting server on port %s", config.AppPort()))

	// START SERVER
	if err := router.Run(fmt.Sprintf(":%s", config.AppPort())); err != nil {
		panic(err)
	}
}
