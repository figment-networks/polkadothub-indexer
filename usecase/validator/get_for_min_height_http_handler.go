package validator

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
		err := errors.New("invalid request parameters")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp, err := h.getUseCase().Execute(req.Height)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *getForMinHeightHttpHandler) getUseCase() *getForMinHeightUseCase {
	if h.useCase == nil {
		return NewGetForMinHeightUseCase(h.syncablesDb, h.validatorAggDb)
	}
	return h.useCase
}
