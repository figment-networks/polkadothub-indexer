package types

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type HttpHandler interface {
	Handle(*gin.Context)
}

type CliHandler interface {
	Handle(*cobra.Command, []string)
}

type JobHandler interface {
	Handle()
}