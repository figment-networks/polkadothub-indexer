package block

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getBlockSummaryHttpHandler)(nil)

	ErrInvalidIntervalPeriod = errors.New("invalid interval and/or period")
)

type getBlockSummaryHttpHandler struct {
	useCase *getBlockSummaryUseCase

	blockSummaryDb store.BlockSummary
}

func NewGetBlockSummaryHttpHandler(blockSummaryDb store.BlockSummary) *getBlockSummaryHttpHandler {
	return &getBlockSummaryHttpHandler{
		blockSummaryDb: blockSummaryDb,
	}
}

type GetBlockTimesForIntervalRequest struct {
	Interval types.SummaryInterval `form:"interval" binding:"required"`
	Period   string                `form:"period" binding:"required"`
}

func (h *getBlockSummaryHttpHandler) Handle(c *gin.Context) {
	req, err := h.validateParams(c)
	if err != nil {
		logger.Error(err)
		http.BadRequest(c, err)
		return
	}

	resp, err := h.getUseCase().Execute(req.Interval, req.Period)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getBlockSummaryHttpHandler) validateParams(c *gin.Context) (*GetBlockTimesForIntervalRequest, error) {
	var req GetBlockTimesForIntervalRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}

	if !req.Interval.Valid() {
		return nil, ErrInvalidIntervalPeriod
	}

	return &req, nil
}

func (h *getBlockSummaryHttpHandler) getUseCase() *getBlockSummaryUseCase {
	if h.useCase == nil {
		return NewGetBlockSummaryUseCase(h.blockSummaryDb)
	}
	return h.useCase
}
