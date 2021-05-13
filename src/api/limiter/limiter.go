package limiter

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const LimitRate = 10
const LimitBurst = 30

type IPRateLimiter struct {
	Visitors map[string]*Visitor
	Mu       *sync.RWMutex
}

type Visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func (rl *IPRateLimiter) Init() {
	go rl.cleanupVisitors()
}

func (rl *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	v, exists := rl.Visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(LimitRate, LimitBurst)
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
			log.Println(err.Error())
			ip = ""
		}

		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText( http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
