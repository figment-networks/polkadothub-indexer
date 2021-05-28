package apr

import (
	"time"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getAprByAddressHttpHandler)(nil)
)

type getAprByAddressHttpHandler struct {
	useCase *getAprByAddressUseCase

	accountDb   store.AccountEraSeq
	rewardDb    store.Rewards
	syncablesDb store.Syncables
}

func NewGetAprByAddressHttpHandler(accountDb store.AccountEraSeq, rewardDb store.Rewards, syncablesDb store.Syncables) *getAprByAddressHttpHandler {
	return &getAprByAddressHttpHandler{
		accountDb:   accountDb,
		rewardDb:    rewardDb,
		syncablesDb: syncablesDb,
	}
}

type queryParams struct {
	Account string    `form:"acount" binding:"required"`
	Start   time.Time `form:"start_time" binding:"required" time_format:"2006-01-02"`
	End     time.Time `form:"end_time" binding:"-" time_format:"2006-01-02"`
}

func (h *getAprByAddressHttpHandler) Handle(c *gin.Context) {
	var params queryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		http.BadRequest(c, errors.New("required start_time or account is missing, and/or start_time end_time must be in format: 2006-01-02"))
		return
	}

	resp, err := h.getUseCase().Execute(params.Account, *types.NewTimeFromTime(params.Start), *types.NewTimeFromTime(params.End))
	if http.ShouldReturn(c, err) {
		return
	}

	http.JsonOK(c, resp)
}

func (h *getAprByAddressHttpHandler) getUseCase() *getAprByAddressUseCase {
	if h.useCase == nil {
		h.useCase = NewGetAprByAddressUseCase(h.accountDb, h.rewardDb, h.syncablesDb)
	}
	return h.useCase
}
