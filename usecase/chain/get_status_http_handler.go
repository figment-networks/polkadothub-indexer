package chain

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
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

// swagger:parameters getStatus
type GetStatusRequest struct {
	// IncludeChainStatus
	//
	// in: query
	IncludeChainStatus bool `form:"include_chain" binding:"-"`
}

func (h *getStatusHttpHandler) Handle(c *gin.Context) {
	var req GetStatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid query"))
		return
	}

	resp, err := h.getUseCase().Execute(req.IncludeChainStatus)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getStatusHttpHandler) getUseCase() *getStatusUseCase {
	if h.useCase == nil {
		return NewGetStatusUseCase(h.client, h.syncablesDb)
	}
	return h.useCase
}
