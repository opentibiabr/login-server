package limiter

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/configs"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	Visitors map[string]*Visitor
	Mu       *sync.RWMutex
	Configs  configs.RateLimiter
}

type Visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func (rl *IPRateLimiter) Init() {
	rl.Configs = configs.GetRateLimiterConfigs()
	go rl.cleanupVisitors()
}

func (rl *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	v, exists := rl.Visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.Configs.Rate, rl.Configs.Burst)
		rl.Visitors[ip] = &Visitor{limiter, time.Now()}
		return limiter
	}

	v.LastSeen = time.Now()
	return v.Limiter
}

func (rl *IPRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		rl.Mu.Lock()
		for ip, v := range rl.Visitors {
			if time.Since(v.LastSeen) > 3*time.Minute {
				delete(rl.Visitors, ip)
			}
		}
		rl.Mu.Unlock()
	}
}

func (rl *IPRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		}
		c.Next()
	}
}
