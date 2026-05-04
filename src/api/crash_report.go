package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (_api *Api) crashReport(c *gin.Context) {
	if c.Request.Body != nil {
		_, _ = io.Copy(io.Discard, c.Request.Body)
	}

	c.Status(http.StatusNoContent)
}
