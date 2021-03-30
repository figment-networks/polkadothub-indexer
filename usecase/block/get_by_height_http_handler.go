package block

import (
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getByHeightHttpHandler)(nil)
)

type getByHeightHttpHandler struct {
	client *client.Client

	useCase *getByHeightUseCase

	syncablesDb store.Syncables
}

func NewGetByHeightHttpHandler(c *client.Client, syncablesDb store.Syncables) *getByHeightHttpHandler {
	return &getByHeightHttpHandler{
		syncablesDb: syncablesDb,
		client:      c,
	}
}

// swagger:parameters getBlockByHeight
type Request struct {
	// Height
	//
	// in: query
	Height *int64 `form:"height" binding:"-"`
}

func (h *getByHeightHttpHandler) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid height"))
		return
	}

	ds, err := h.getUseCase().Execute(req.Height)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, ds)
}

func (h *getByHeightHttpHandler) getUseCase() *getByHeightUseCase {
	if h.useCase == nil {
		return NewGetByHeightUseCase(h.client, h.syncablesDb)
	}
	return h.useCase
}
