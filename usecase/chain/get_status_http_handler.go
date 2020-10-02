package chain

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
)

var (
	_ types.HttpHandler = (*getStatusHttpHandler)(nil)
)

type getStatusHttpHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getStatusUseCase
}

func NewGetStatusHttpHandler(db *store.Store, client *client.Client) *getStatusHttpHandler {
	return &getStatusHttpHandler{
		db:     db,
		client: client,
	}
}

func (h *getStatusHttpHandler) Handle(c *gin.Context) {
	resp, err := h.getUseCase().Execute(c)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, resp)
}

func (h *getStatusHttpHandler) getUseCase() *getStatusUseCase {
	if h.useCase == nil {
		return NewGetStatusUseCase(h.db, h.client)
	}
	return h.useCase
}
