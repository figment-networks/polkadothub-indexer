package validator

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getForMinHeightHttpHandler)(nil)
)

type getForMinHeightHttpHandler struct {
	useCase *getForMinHeightUseCase

	syncablesDb    store.Syncables
	validatorAggDb store.ValidatorAgg
}

func NewGetForMinHeightHttpHandler(syncablesDb store.Syncables, validatorAggDb store.ValidatorAgg) *getForMinHeightHttpHandler {
	return &getForMinHeightHttpHandler{
		syncablesDb:    syncablesDb,
		validatorAggDb: validatorAggDb,
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
		return NewGetForMinHeightUseCase(h.syncablesDb, h.validatorAggDb)
	}
	return h.useCase
}
