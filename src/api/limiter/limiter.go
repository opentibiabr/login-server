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
	// KeyForRequest may partition server-defined request classes for the same IP.
	KeyForRequest func(*gin.Context, string) (string, bool)
}

type Visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func (rl *IPRateLimiter) Init() {
	rl.Configs = configs.GetRateLimiterConfigs()
	go rl.cleanupVisitors()
}

// The caller holds Mu, including while probing or consuming a token.
func (rl *IPRateLimiter) getVisitorLocked(ip string) *rate.Limiter {
	v, exists := rl.Visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.Configs.Rate, rl.Configs.Burst)
		rl.Visitors[ip] = &Visitor{limiter, time.Now()}
		return limiter
	}

	v.LastSeen = time.Now()
	return v.Limiter
}

// CanAllow checks capacity without consuming a token. Probe and final admission
// share Mu so a canceled probe cannot interfere with another request's Allow.
func (rl *IPRateLimiter) CanAllow(key string) bool {
	return rl.canAllowAt(key, time.Now())
}

// Observe both candidate classes at one instant under the same ownership lock.
func (rl *IPRateLimiter) CanAllowAny(firstKey, secondKey string) bool {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()
	now := time.Now()
	return rl.canAllowLocked(firstKey, now) || rl.canAllowLocked(secondKey, now)
}

func (rl *IPRateLimiter) canAllowAt(key string, now time.Time) bool {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()
	return rl.canAllowLocked(key, now)
}

func (rl *IPRateLimiter) canAllowLocked(key string, now time.Time) bool {
	limiter := rl.getVisitorLocked(key)
	limit, burst := limiter.Limit(), limiter.Burst()
	if limit == rate.Inf {
		return true
	}
	if burst < 1 {
		return false
	}
	// The pinned rate package admits nonpositive rates with a positive burst;
	// its reservation time can precede now, in which case CancelAt does nothing.
	if limit <= 0 {
		return true
	}
	reservation := limiter.ReserveN(now, 1)
	allowed := reservation.OK() && reservation.DelayFrom(now) == 0
	reservation.CancelAt(now)
	return allowed
}

func (rl *IPRateLimiter) allow(key string) bool {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()
	return rl.getVisitorLocked(key).Allow()
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
		possible := true
		if rl.KeyForRequest != nil {
			ip, possible = rl.KeyForRequest(c, ip)
		}

		if !possible || !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
