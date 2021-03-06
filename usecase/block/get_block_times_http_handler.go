package block

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/log"
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

// swagger:parameters getBlockTimes
type GetBlockTimesRequest struct {
	// Limit
	//
	// required: true
	// in: path
	Limit int64 `json:"limit" uri:"limit" binding:"required"`
}

func (h *getBlockTimesHttpHandler) Handle(c *gin.Context) {
	var req GetBlockTimesRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.Error(err)
		http.BadRequest(c, errors.New("invalid height"))
		return
	}

	resp := h.getUseCase().Execute(req.Limit)
	http.JsonOK(c, resp)
}

func (h *getBlockTimesHttpHandler) getUseCase() *getBlockTimesUseCase {
	if h.useCase == nil {
		return NewGetBlockTimesUseCase(h.blockSeqDb)
	}
	return h.useCase
}
