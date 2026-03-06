package mid

import (
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 3)
	defer rl.Stop()

	// First 3 requests (burst) should be allowed.
	for i := 0; i < 3; i++ {
		if !rl.Allow("10.0.0.1") {
			t.Fatalf("request %d should be allowed (within burst)", i+1)
		}
	}
}

func TestRateLimiter_ExceedBurst(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 3)
	defer rl.Stop()

	// Exhaust burst.
	for i := 0; i < 3; i++ {
		rl.Allow("10.0.0.2")
	}

	// Burst+1 should be denied.
	if rl.Allow("10.0.0.2") {
		t.Fatal("burst+1 request should be denied")
	}
}

func TestRateLimiter_PerIP(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 1)
	defer rl.Stop()

	// Exhaust burst for IP A.
	rl.Allow("10.0.0.3")
	if rl.Allow("10.0.0.3") {
		t.Fatal("second request from IP A should be denied")
	}

	// IP B should have its own independent limit.
	if !rl.Allow("10.0.0.4") {
		t.Fatal("first request from IP B should be allowed")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 5)

	done := make(chan struct{})
	go func() {
		rl.Stop()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("Stop did not return within 1 second")
	}
}

func TestIPExtractor_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.100:54321"

	ip := RemoteAddrExtractor(r)
	if ip != "192.168.1.100" {
		t.Fatalf("expected 192.168.1.100, got %s", ip)
	}
}

func TestTrustedProxyExtractor(t *testing.T) {
	cidrs := ParseTrustedCIDRs("10.0.0.0/8")
	extract := TrustedProxyExtractor(cidrs)

	// Request comes directly from client (not a trusted proxy) — XFF ignored.
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.1:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")

	ip := extract(r)
	if ip != "203.0.113.1" {
		t.Fatalf("expected RemoteAddr 203.0.113.1 (untrusted proxy), got %s", ip)
	}

	// Request comes from trusted proxy — use rightmost untrusted XFF IP.
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.RemoteAddr = "10.0.0.1:1234" // trusted proxy
	r2.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.2")

	ip2 := extract(r2)
	if ip2 != "203.0.113.5" {
		t.Fatalf("expected rightmost untrusted IP 203.0.113.5, got %s", ip2)
	}
}
