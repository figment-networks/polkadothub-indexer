package block

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

var (
	_ types.HttpHandler = (*getBlockTimesHttpHandler)(nil)
)

type getBlockTimesHttpHandler struct {
	db     store.Store
	client *client.Client

	useCase *getBlockTimesUseCase
}

func NewGetBlockTimesHttpHandler(db store.Store, client *client.Client) *getBlockTimesHttpHandler {
	return &getBlockTimesHttpHandler{
		db:     db,
		client: client,
	}
}

type GetBlockTimesRequest struct {
	Limit int64 `uri:"limit" binding:"required"`
}

func (h *getBlockTimesHttpHandler) Handle(c *gin.Context) {
	var req GetBlockTimesRequest
	if err := c.ShouldBindUri(&req); err != nil {
		//log.Error(err)
		err := errors.New("invalid height")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp := h.getUseCase().Execute(req.Limit)

	c.JSON(http.StatusOK, resp)
}

func (h *getBlockTimesHttpHandler) getUseCase() *getBlockTimesUseCase {
	if h.useCase == nil {
		return NewGetBlockTimesUseCase(h.db)
	}
	return h.useCase
}
