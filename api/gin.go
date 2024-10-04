package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Gin handler adapter
func GinHandlerFunc(handlerFunc func(http.ResponseWriter, *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerFunc(c.Writer, c.Request)
	}
}
