package api

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/logger"
)

// newRouter builds the HTTP router with an explicit proxy trust list.
//
// gin trusts every peer by default, so c.ClientIP() returns whatever the client
// puts in X-Forwarded-For or X-Real-IP. The per-IP rate limiter and the request
// log both key on c.ClientIP(), which let any client pick a fresh "IP" for each
// login attempt. Forwarded headers are now honoured only from the peers listed
// in LOGIN_TRUSTED_PROXIES (comma-separated IPs or CIDRs); by default none are
// trusted and the TCP peer address is used.
func newRouter(trustedProxies string) *gin.Engine {
	router := gin.New()
	proxies := parseTrustedProxies(trustedProxies)
	if err := router.SetTrustedProxies(proxies); err != nil {
		logger.Error(fmt.Errorf("invalid LOGIN_TRUSTED_PROXIES, trusting no proxy: %v", err))
		_ = router.SetTrustedProxies(nil)
	}
	return router
}

func parseTrustedProxies(value string) []string {
	var proxies []string
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			proxies = append(proxies, part)
		}
	}
	return proxies
}
