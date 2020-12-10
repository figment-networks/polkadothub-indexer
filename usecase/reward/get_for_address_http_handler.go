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

type GetForStashAccountRequest struct {
	StashAccount string `uri:"stash_account" binding:"required"`
	Start        int64  `form:"start" binding:"-"`
	End          int64  `form:"end" binding:"-"`
}

func (h *getForStashAccountHttpHandler) Handle(c *gin.Context) {
	var req GetForStashAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid stash account"))
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid start or/and end"))
		return
	}

	resp, err := h.getUseCase().Execute(req.StashAccount, req.Start, req.End)
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
