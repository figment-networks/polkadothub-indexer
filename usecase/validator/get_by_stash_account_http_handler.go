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
	_ types.HttpHandler = (*getByStashAccountHttpHandler)(nil)
)

type getByStashAccountHttpHandler struct {
	useCase *getByStashAccountUseCase

	accountEraSeqDb store.AccountEraSeq
	validatorDb     store.Validators
}

func NewGetByStashAccountHttpHandler(accountEraSeqDb store.AccountEraSeq, validatorDb store.Validators) *getByStashAccountHttpHandler {
	return &getByStashAccountHttpHandler{
		accountEraSeqDb: accountEraSeqDb,
		validatorDb:     validatorDb,
	}
}

type GetByEntityUidRequest struct {
	StashAccount  string `uri:"stash_account" binding:"required"`
	SessionsLimit int64  `form:"sessions_limit" binding:"-"`
	ErasLimit     int64  `form:"eras_limit" binding:"-"`
}

func (h *getByStashAccountHttpHandler) Handle(c *gin.Context) {
	var req GetByEntityUidRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid stash account"))
		return
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid sequences limit"))
		return
	}

	resp, err := h.getUseCase().Execute(req.StashAccount, req.SessionsLimit, req.ErasLimit)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getByStashAccountHttpHandler) getUseCase() *getByStashAccountUseCase {
	if h.useCase == nil {
		return NewGetByStashAccountUseCase(h.accountEraSeqDb, h.validatorDb)

	}
	return h.useCase
}
