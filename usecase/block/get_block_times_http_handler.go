package block

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getBlockTimesHttpHandler)(nil)
)

type getBlockTimesHttpHandler struct {
	useCase *getBlockTimesUseCase

	blockSeqDb store.BlockSeq
}

func NewGetBlockTimesHttpHandler(blockSeqDb store.BlockSeq) *getBlockTimesHttpHandler {
	return &getBlockTimesHttpHandler{
		blockSeqDb: blockSeqDb,
	}
}

type GetBlockTimesRequest struct {
	Limit int64 `uri:"limit" binding:"required"`
}

func (h *getBlockTimesHttpHandler) Handle(c *gin.Context) {
	var req GetBlockTimesRequest
	if err := c.ShouldBindUri(&req); err != nil {
		err := errors.New("invalid height")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp := h.getUseCase().Execute(req.Limit)

	c.JSON(http.StatusOK, resp)
}

func (h *getBlockTimesHttpHandler) getUseCase() *getBlockTimesUseCase {
	if h.useCase == nil {
		return NewGetBlockTimesUseCase(h.blockSeqDb)
	}
	return h.useCase
}
