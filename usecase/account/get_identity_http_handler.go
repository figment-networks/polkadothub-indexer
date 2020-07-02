package account

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
	_ types.HttpHandler = (*getIdentityHttpHandler)(nil)
)

type getIdentityHttpHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getIdentityUseCase
}

func NewGetIdentityHttpHandler(db *store.Store, c *client.Client) *getIdentityHttpHandler {
	return &getIdentityHttpHandler{
		db: db,
		client: c,
	}
}

type GetIdentityRequest struct {
	Address string `uri:"address" binding:"required"`
}

func (h *getIdentityHttpHandler) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid address")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ds, err := h.getUseCase().Execute(req.Address)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *getIdentityHttpHandler) getUseCase() *getIdentityUseCase {
	if h.useCase == nil {
		return NewGetIdentityUseCase(h.db, h.client)
	}
	return h.useCase
}



