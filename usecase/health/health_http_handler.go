package health

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	_ types.HttpHandler = (*httpHealthHandler)(nil)
)

type httpHealthHandler struct {}

func NewHealthHttpHandler() *httpHealthHandler {
	return &httpHealthHandler{}
}

func (h *httpHealthHandler) Handle(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}