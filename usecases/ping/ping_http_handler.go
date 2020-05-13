package ping

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HttpHandler interface {
	Ping(*gin.Context)
}

type httpHandler struct {}

func NewHttpHandler() types.HttpHandler {
	return &httpHandler{}
}

func (h *httpHandler) Handle(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}