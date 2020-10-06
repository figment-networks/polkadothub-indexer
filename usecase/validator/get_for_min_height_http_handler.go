package validator

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getForMinHeightHttpHandler)(nil)
)

type getForMinHeightHttpHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getForMinHeightUseCase
}

func NewGetForMinHeightHttpHandler(db *store.Store, c *client.Client) *getForMinHeightHttpHandler {
	return &getForMinHeightHttpHandler{
		db:     db,
		client: c,
	}
}

type GetForMinHeightRequest struct {
	Height *int64 `uri:"height" binding:"required"`
}

func (h *getForMinHeightHttpHandler) Handle(c *gin.Context) {
	var req GetForMinHeightRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid request parameters"))
		return
	}

	resp, err := h.getUseCase().Execute(req.Height)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getForMinHeightHttpHandler) getUseCase() *getForMinHeightUseCase {
	if h.useCase == nil {
		return NewGetForMinHeightUseCase(h.db)
	}
	return h.useCase
}
