package reward

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getForStashAccountHttpHandler)(nil)
)

type getForStashAccountHttpHandler struct {
	rewardDb store.Rewards

	useCase *getForStashAccountUseCase
}

func NewGetForStashAccountHttpHandler(rewardDb store.Rewards) *getForStashAccountHttpHandler {
	return &getForStashAccountHttpHandler{
		rewardDb: rewardDb,
	}
}

// swagger:parameters getRewards
type uriParams struct {
	// StashAccount
	//
	// required: true
	// in: path
	StashAccount string `json:"stash_account" uri:"stash_account" binding:"required"`
}

// swagger:parameters getRewards
type queryParams struct {
	// Start
	//
	// in: query
	Start int64 `json:"start"  form:"start" binding:"-"`
	// End
	//
	// in: query
	End int64 `json:"end"  form:"end" binding:"-"`
	// ValidatorStash
	//
	// in: query
	ValidatorStash string `json:"validator" form:"validator" binding:"-"`
}

func (h *getForStashAccountHttpHandler) Handle(c *gin.Context) {
	var uri uriParams
	if err := c.ShouldBindUri(&uri); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid stash account"))
		return
	}
	var query queryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid start and/or end and/or validator"))
		return
	}

	resp, err := h.getUseCase().Execute(uri.StashAccount, query.Start, query.End, query.ValidatorStash)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getForStashAccountHttpHandler) getUseCase() *getForStashAccountUseCase {
	if h.useCase == nil {
		h.useCase = NewGetForStashAccountUseCase(h.rewardDb)
	}
	return h.useCase
}
