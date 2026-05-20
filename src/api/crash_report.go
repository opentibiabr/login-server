package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxCrashReportBodyBytes = 1 << 20 // 1 MiB

func (_api *Api) crashReport(c *gin.Context) {
	if c.Request.Body != nil {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCrashReportBodyBytes)
		if _, err := io.Copy(io.Discard, c.Request.Body); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				c.Status(http.StatusRequestEntityTooLarge)
				return
			}

			c.Status(http.StatusBadRequest)
			return
		}
	}

	c.Status(http.StatusNoContent)
}
