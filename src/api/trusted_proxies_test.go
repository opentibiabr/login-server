package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func clientIPFor(router *gin.Engine, remoteAddr string, headers map[string]string) string {
	router.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Body.String()
}

func TestClientIPIgnoresForwardedHeadersByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, h := range []string{"X-Forwarded-For", "X-Real-IP"} {
		ip := clientIPFor(newRouter(""), "203.0.113.9:5555", map[string]string{h: "198.51.100.7"})
		assert.Equal(t, "203.0.113.9", ip, h)
	}
}

func TestClientIPHonoursConfiguredProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter("172.30.0.0/16")
	ip := clientIPFor(router, "172.30.0.5:5555", map[string]string{"X-Forwarded-For": "198.51.100.7"})
	assert.Equal(t, "198.51.100.7", ip)
}

func TestClientIPIgnoresHeadersFromUntrustedPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter("172.30.0.0/16")
	ip := clientIPFor(router, "203.0.113.9:5555", map[string]string{"X-Forwarded-For": "198.51.100.7"})
	assert.Equal(t, "203.0.113.9", ip)
}

func TestInvalidTrustedProxiesTrustsNone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ip := clientIPFor(newRouter("not-an-ip"), "203.0.113.9:5555", map[string]string{"X-Forwarded-For": "198.51.100.7"})
	assert.Equal(t, "203.0.113.9", ip)
}

func TestParseTrustedProxies(t *testing.T) {
	assert.Nil(t, parseTrustedProxies(""))
	assert.Equal(t, []string{"10.0.0.1", "172.30.0.0/16"}, parseTrustedProxies(" 10.0.0.1 , ,172.30.0.0/16 "))
}
