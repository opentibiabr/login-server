package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCrashReportReturnsNoContentWithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/crash-report", (&Api{}).crashReport)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/crash-report", bytes.NewReader(bytes.Repeat([]byte("a"), 1024)))
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestCrashReportReturnsPayloadTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/crash-report", (&Api{}).crashReport)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/crash-report", bytes.NewReader(bytes.Repeat([]byte("a"), maxCrashReportBodyBytes+1)))
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}
