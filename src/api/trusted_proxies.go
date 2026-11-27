package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
)

// newRouter builds the HTTP router with an explicit proxy trust list.
//
// gin trusts every peer by default, so c.ClientIP() returns whatever the client
// puts in X-Forwarded-For or X-Real-IP. The per-IP rate limiter and the request
// log both key on c.ClientIP(), which let any client pick a fresh "IP" for each
// login attempt. Forwarded headers are now honoured only from the peers listed
// in LOGIN_TRUSTED_PROXIES (comma-separated IPs or CIDRs, read once by
// configs.getTrustedProxies, the same list the X-Forwarded-Proto check uses); by
// default none are trusted and the TCP peer address is used.
func newRouter(proxies []string) *gin.Engine {
	router := gin.New()
	if err := router.SetTrustedProxies(proxies); err != nil {
		logger.Error(fmt.Errorf("invalid %s (%v): the whole list is ignored, no proxy is trusted and every request is attributed to the TCP peer address", configs.EnvLoginTrustedProxiesKey, err))
		_ = router.SetTrustedProxies(nil)
		return router
	}
	if len(proxies) == 0 {
		logger.Info(fmt.Sprintf("%s is empty: forwarded client IP headers are ignored. Behind a reverse proxy, set it to the proxy address, otherwise all clients share one rate limit bucket.", configs.EnvLoginTrustedProxiesKey))
	}
	return router
}
