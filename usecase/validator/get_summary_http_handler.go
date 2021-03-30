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
	_ types.HttpHandler = (*getSummaryHttpHandler)(nil)

	ErrInvalidIntervalPeriod = errors.New("invalid interval and/or period")
)

type getSummaryHttpHandler struct {
	useCase *getSummaryUseCase

	syncablesDb        store.Syncables
	validatorSummaryDb store.ValidatorSummary
}

func NewGetSummaryHttpHandler(syncablesDb store.Syncables, validatorSummaryDb store.ValidatorSummary) *getSummaryHttpHandler {
	return &getSummaryHttpHandler{
		syncablesDb:        syncablesDb,
		validatorSummaryDb: validatorSummaryDb,
	}
}

// swagger:parameters getValidatorSummary
type GetSummaryRequest struct {
	// Interval
	//
	// required: true
	// in: query
	Interval types.SummaryInterval `json:"interval" form:"interval" binding:"required"`
	// Period
	//
	// required: true
	// in: query
	Period string `json:"period" form:"period" binding:"required"`
	// StashAccount
	//
	// required: true
	// in: query
	StashAccount string `json:"stash_account" form:"stash_account" binding:"-"`
}

func (h *getSummaryHttpHandler) Handle(c *gin.Context) {
	req, err := h.validateParams(c)
	if err != nil {
		logger.Error(err)
		http.BadRequest(c, err)
		return
	}

	resp, err := h.getUseCase().Execute(req.Interval, req.Period, req.StashAccount)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getSummaryHttpHandler) validateParams(c *gin.Context) (*GetSummaryRequest, error) {
	var req GetSummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}

	if !req.Interval.Valid() {
		return nil, ErrInvalidIntervalPeriod
	}

	return &req, nil
}

func (h *getSummaryHttpHandler) getUseCase() *getSummaryUseCase {
	if h.useCase == nil {
		return NewGetSummaryUseCase(h.syncablesDb, h.validatorSummaryDb)
	}
	return h.useCase
}
