package auth

import (
	"testing"
	"time"
)

func TestBlocklist_AddAndRevoke(t *testing.T) {
	bl := NewBlocklist()
	defer bl.Stop()

	jti := "test-jti-001"
	bl.Add(jti, time.Now().Add(5*time.Minute))

	if !bl.IsRevoked(jti) {
		t.Fatal("expected jti to be revoked")
	}
}

func TestBlocklist_ExpiredNotRevoked(t *testing.T) {
	bl := NewBlocklist()
	defer bl.Stop()

	jti := "test-jti-expired"
	bl.Add(jti, time.Now().Add(-1*time.Second)) // already expired

	if bl.IsRevoked(jti) {
		t.Fatal("expected expired jti to not be revoked")
	}
}

func TestBlocklist_EmptyJTI(t *testing.T) {
	bl := NewBlocklist()
	defer bl.Stop()

	bl.Add("", time.Now().Add(5*time.Minute)) // no-op

	if bl.IsRevoked("") {
		t.Fatal("empty jti should never be revoked")
	}
}

func TestBlocklist_Stop(t *testing.T) {
	bl := NewBlocklist()

	// Stop should return without blocking or panicking.
	done := make(chan struct{})
	go func() {
		bl.Stop()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("Stop did not return within 1 second")
	}
}
