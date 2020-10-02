package http

import (
	"net/http"

	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/gin-gonic/gin"
)

// BadRequest renders a HTTP 400 bad request response
func BadRequest(c *gin.Context, err interface{}) {
	jsonError(c, http.StatusBadRequest, err)
}

// NotFound renders a HTTP 404 not found response
func NotFound(c *gin.Context, err interface{}) {
	jsonError(c, http.StatusNotFound, err)
}

// ServerError renders a HTTP 500 error response
func ServerError(c *gin.Context, err interface{}) {
	jsonError(c, http.StatusInternalServerError, err)
}

// JsonOK renders a successful response
func JsonOK(c *gin.Context, data interface{}) {
	switch data.(type) {
	case []byte:
		c.Header("Content-Type", "application/json")
		c.String(200, "%s", data)
	default:
		c.JSON(200, data)
	}
}

// jsonError renders an error response
func jsonError(c *gin.Context, status int, err interface{}) {
	var message interface{}

	switch v := err.(type) {
	case error:
		message = v.Error()
	default:
		message = v
	}

	c.AbortWithStatusJSON(status, gin.H{
		"status": status,
		"error":  message,
	})
}

// ShouldReturn is a shorthand method for handling resource errors
func ShouldReturn(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if err == store.ErrNotFound {
		NotFound(c, err)
	} else {
		ServerError(c, err)
	}

	return true
}
