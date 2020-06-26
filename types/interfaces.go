package types

import (
	"github.com/gin-gonic/gin"
)

type HttpHandler interface {
	Handle(*gin.Context)
}

type WorkerHandler interface {
	Handle()
}
