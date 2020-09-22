package account

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getDetailsHttpHandler)(nil)
)

type getDetailsHttpHandler struct {
	client *client.Client

	useCase *getDetailsUseCase

	accountEraSeqDb store.AccountEraSeq
	eventSeqDb      store.EventSeq
}

func NewGetDetailsHttpHandler(c *client.Client, accountEraSeqDb store.AccountEraSeq, eventSeqDb store.EventSeq) *getDetailsHttpHandler {
	return &getDetailsHttpHandler{
		client: c,

		accountEraSeqDb: accountEraSeqDb,
		eventSeqDb:      eventSeqDb,
	}
}

type GetDetailsRequest struct {
	StashAccount string `uri:"stash_account" binding:"required"`
}

func (h *getDetailsHttpHandler) Handle(c *gin.Context) {
	var req GetDetailsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		err := errors.New("invalid stash account")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ds, err := h.getUseCase().Execute(req.StashAccount)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *getDetailsHttpHandler) getUseCase() *getDetailsUseCase {
	if h.useCase == nil {
		return NewGetDetailsUseCase(h.client, h.accountEraSeqDb, h.eventSeqDb)
	}
	return h.useCase
}
