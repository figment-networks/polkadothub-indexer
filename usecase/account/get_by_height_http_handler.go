package account

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
	_ types.HttpHandler = (*getByHeightHttpHandler)(nil)
)

type getByHeightHttpHandler struct {
	client *client.Client

	useCase *getByHeightUseCase

	syncablesDb store.FindMostRecenter
}

func NewGetByHeightHttpHandler(c *client.Client, syncablesDb store.FindMostRecenter) *getByHeightHttpHandler {
	return &getByHeightHttpHandler{
		client:      c,
		syncablesDb: syncablesDb,
	}
}

// swagger:parameters getAccountByHeight
type GetByHeightRequest struct {
	// StashAccount
	//
	// required: true
	// in: path
	StashAccount string `json:"stash_account" uri:"stash_account" binding:"required"`
	// Height of block
	//
	// in: query
	Height *int64 `json:"height" form:"height" binding:"-"`
}

func (h *getByHeightHttpHandler) Handle(c *gin.Context) {
	var req GetByHeightRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid stash account"))
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid height"))
		return
	}

	ds, err := h.getUseCase().Execute(req.StashAccount, req.Height)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, ds)
}

func (h *getByHeightHttpHandler) getUseCase() *getByHeightUseCase {
	if h.useCase == nil {
		return NewGetByHeightUseCase(h.client, h.syncablesDb)
	}
	return h.useCase
}
