package tcpprint_test

import (
	"context"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/tcpprint"
)

// Test_TCPPrinter_EmptyHostNoop verifies that an empty hostPort short-
// circuits SendZPL with nil — honouring mux.Config's "Empty
// PrinterHostPort disables actual network dispatch" contract. Without
// the guard, an empty dial target would be rejected by the kernel at
// request time in environments that intentionally leave it unset.
func Test_TCPPrinter_EmptyHostNoop(t *testing.T) {
	p := tcpprint.New("", 500*time.Millisecond)
	if err := p.SendZPL(context.Background(), []byte("^XA^XZ")); err != nil {
		t.Fatalf("SendZPL with empty host should be no-op, got: %v", err)
	}
}
