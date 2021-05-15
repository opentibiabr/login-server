package limiter

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"time"

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

func (rl *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			logger.Error(err)
			ip = ""
		}

		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
			logger.WithFields(logrus.Fields{"ip": ip}).Info("too many requests")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
