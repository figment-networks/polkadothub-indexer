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
	_ types.HttpHandler = (*getByHeightHttpHandler)(nil)
)

type getByHeightHttpHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getByHeightUseCase
}

func NewGetByHeightHttpHandler(db *store.Store, c *client.Client) *getByHeightHttpHandler {
	return &getByHeightHttpHandler{
		db: db,
		client: c,
	}
}

type Request struct {
	Address string `uri:"address" binding:"required"`
	Height *int64 `form:"height" binding:"-"`
}

func (h *getByHeightHttpHandler) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid stash account")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid height")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ds, err := h.getUseCase().Execute(req.Address, req.Height)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *getByHeightHttpHandler) getUseCase() *getByHeightUseCase {
	if h.useCase == nil {
		return NewGetByHeightUseCase(h.db, h.client)
	}
	return h.useCase
}



