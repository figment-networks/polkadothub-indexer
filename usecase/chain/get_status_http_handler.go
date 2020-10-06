package chain

import (
	"errors"
	"net/http"

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
	db     *store.Store
	client *client.Client

	useCase *getStatusUseCase
}

func NewGetStatusHttpHandler(db *store.Store, client *client.Client) *getStatusHttpHandler {
	return &getStatusHttpHandler{
		db:     db,
		client: client,
	}
}

type GetStatusRequest struct {
	IncludeChainStatus bool `form:"include_chain" binding:"-"`
}

func (h *getStatusHttpHandler) Handle(c *gin.Context) {
	var req GetStatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid query")
		c.JSON(http.StatusBadRequest, err)
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
		return NewGetStatusUseCase(h.db, h.client)
	}
	return h.useCase
}
