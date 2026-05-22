package database

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

func writeServiceError(c *gin.Context, err *serviceerrors.PublicError) {
	if err.Cause != nil {
		logger.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode":    err.Code,
		"errorMessage": err.Message,
	})
}
