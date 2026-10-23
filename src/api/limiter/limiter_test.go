package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/opentibiabr/login-server/src/configs"
	"golang.org/x/time/rate"
)

func testLimiter(limit rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{Visitors: make(map[string]*Visitor), Mu: &sync.RWMutex{}, Configs: configs.RateLimiter{Rate: limit, Burst: burst}}
}

func TestCapacityProbeDoesNotConsumeBurst(t *testing.T) {
	rl := testLimiter(1.0/86400, 12)
	for attempt := 0; attempt < 100; attempt++ {
		if !rl.CanAllow("ip") {
			t.Fatal("capacity probe consumed the burst")
		}
	}
	for attempt := 0; attempt < 12; attempt++ {
		if !rl.allow("ip") {
			t.Fatalf("capacity probe consumed token %d", attempt)
		}
	}
	if rl.CanAllow("ip") || rl.allow("ip") {
		t.Fatal("probe or admission exceeded the burst")
	}
	if !rl.CanAllow("public:ip") {
		t.Fatal("missing class bucket inherited another class's exhaustion")
	}
	// Cleanup may remove an idle bucket; its next use must receive a fresh burst.
	rl.Mu.Lock()
	delete(rl.Visitors, "ip")
	rl.Mu.Unlock()
	if !rl.CanAllow("ip") || !rl.allow("ip") {
		t.Fatal("recreated bucket did not receive its configured burst")
	}
}

func TestCapacityProbeRefillBoundary(t *testing.T) {
	rl := testLimiter(2, 12)
	origin := time.Unix(100, 0)
	rl.Mu.Lock()
	bucket := rl.getVisitorLocked("ip")
	if !bucket.AllowN(origin, 12) {
		t.Fatal("could not exhaust burst fixture")
	}
	rl.Mu.Unlock()
	for _, elapsed := range []time.Duration{100, 200, 300, 400, 499} {
		for probe := 0; probe < 10; probe++ {
			if rl.canAllowAt("ip", origin.Add(elapsed*time.Millisecond)) {
				t.Fatal("fractional token admitted before refill boundary")
			}
		}
	}
	boundary := origin.Add(500 * time.Millisecond)
	for attempt := 0; attempt < 2; attempt++ {
		if !rl.canAllowAt("ip", boundary) {
			t.Fatal("probe lost the token available at refill boundary")
		}
	}
	rl.Mu.Lock()
	admitted := bucket.AllowN(boundary, 1)
	rl.Mu.Unlock()
	if !admitted || rl.canAllowAt("ip", boundary) {
		t.Fatal("probe did not preserve exactly one refilled token")
	}
}

func TestCapacityProbeSpecialLimitsMatchAdmission(t *testing.T) {
	for _, limit := range []rate.Limit{-1, 0, 2, rate.Inf} {
		for _, burst := range []int{-1, 0, 1, 12} {
			rl := testLimiter(limit, burst)
			for attempt := 0; attempt < 20; attempt++ {
				possible := rl.CanAllow("ip")
				actual := rl.allow("ip")
				if possible != actual {
					t.Fatalf("limit %v burst %d: probe %v, admission %v", limit, burst, possible, actual)
				}
			}
		}
	}
}

func TestCapacityProbeConcurrentAdmissionDoesNotChangeQuota(t *testing.T) {
	rl := testLimiter(1.0/86400, 12)
	var admitted int32
	var workers sync.WaitGroup
	for worker := 0; worker < 32; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for attempt := 0; attempt < 20; attempt++ {
				_ = rl.CanAllow("ip")
				if rl.allow("ip") {
					atomic.AddInt32(&admitted, 1)
				}
			}
		}()
	}
	workers.Wait()
	if admitted != 12 {
		t.Fatalf("concurrent probes changed quota: admitted %d, expected 12", admitted)
	}
}

func TestCapacityProbeRefreshesRejectedBucketLastSeen(t *testing.T) {
	rl := testLimiter(2, 0)
	if rl.CanAllow("ip") {
		t.Fatal("zero-burst bucket should reject")
	}
	rl.Mu.Lock()
	rl.Visitors["ip"].LastSeen = time.Now().Add(-4 * time.Minute)
	rl.Mu.Unlock()
	before := time.Now()
	if rl.CanAllow("ip") {
		t.Fatal("zero-burst bucket should still reject")
	}
	if rl.Visitors["ip"].LastSeen.Before(before) {
		t.Fatal("rejected probe did not refresh activity")
	}
}
