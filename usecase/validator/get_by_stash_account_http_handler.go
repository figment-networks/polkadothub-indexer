package validator

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

var (
	_ types.HttpHandler = (*getByStashAccountHttpHandler)(nil)
)

type getByStashAccountHttpHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getByStashAccountUseCase
}

func NewGetByStashAccountHttpHandler(db *store.Store, c *client.Client) *getByStashAccountHttpHandler {
	return &getByStashAccountHttpHandler{
		db:     db,
		client: c,
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
		err := errors.New("invalid stash account")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid sequences limit")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp, err := h.getUseCase().Execute(req.StashAccount, req.SessionsLimit, req.ErasLimit)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *getByStashAccountHttpHandler) getUseCase() *getByStashAccountUseCase {
	if h.useCase == nil {
		return NewGetByStashAccountUseCase(h.db)
	}
	return h.useCase
}