package chain

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getStatusHttpHandler)(nil)
)

type getStatusHttpHandler struct {
	client *client.Client

	useCase *getStatusUseCase

	syncablesDb store.Syncables
}

func NewGetStatusHttpHandler(client *client.Client, syncablesDb store.Syncables) *getStatusHttpHandler {
	return &getStatusHttpHandler{
		client:      client,
		syncablesDb: syncablesDb,
	}
}

func (h *getStatusHttpHandler) Handle(c *gin.Context) {
	resp, err := h.getUseCase().Execute(c)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *getStatusHttpHandler) getUseCase() *getStatusUseCase {
	if h.useCase == nil {
		return NewGetStatusUseCase(h.client, h.syncablesDb)
	}
	return h.useCase
}
