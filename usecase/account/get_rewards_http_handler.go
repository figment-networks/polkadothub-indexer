package account

import (
	"errors"
	"time"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getRewardsHttpHandler)(nil)
)

type getRewardsHttpHandler struct {
	useCase *getRewardsUseCase

	accountEraSeqDb store.AccountEraSeq
	eventSeqDb      store.EventSeq
	syncablesDb     store.Syncables
}

func NewGetRewardsHttpHandler(eventSeqDb store.EventSeq, syncablesDb store.Syncables) *getRewardsHttpHandler {
	return &getRewardsHttpHandler{
		eventSeqDb:  eventSeqDb,
		syncablesDb: syncablesDb,
	}
}

// swagger:parameters getAccountRewards
type uriParams struct {
	// Address
	//
	// required: true
	// in: path
	Address string `json:"stash_account" uri:"stash_account" binding:"required"`
}

// swagger:parameters getAccountRewards
type queryParams struct {
	// Start
	//
	// required: true
	// in: query
	// example: 2006-01-02 15:04:05
	Start time.Time `json:"start" form:"start" binding:"required" time_format:"2006-01-02 15:04:05"`
	// End
	//
	// in: query
	// example: 2006-01-02 15:04:05
	End time.Time `json:"end" form:"end" binding:"-" time_format:"2006-01-02 15:04:05"`
}

func (h *getRewardsHttpHandler) Handle(c *gin.Context) {
	var uri uriParams
	if err := c.ShouldBindUri(&uri); err != nil {
		http.BadRequest(c, errors.New("invalid address"))
		return
	}

	var params queryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		http.BadRequest(c, errors.New("invalid start and/or end params: must be in format \"2006-01-02 15:04:05\""))
		return
	}

	resp, err := h.getUseCase().Execute(uri.Address, params.Start, params.End)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getRewardsHttpHandler) getUseCase() *getRewardsUseCase {
	if h.useCase == nil {
		return NewGetRewardsUseCase(h.eventSeqDb, h.syncablesDb)
	}
	return h.useCase
}
