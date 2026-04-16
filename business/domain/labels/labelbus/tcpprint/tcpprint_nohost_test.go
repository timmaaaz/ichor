package tcpprint_test

import (
	"context"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/tcpprint"
)

// Test_TCPPrinter_EmptyHostNoop verifies that an empty host short-
// circuits SendZPL with nil — honouring mux.Config's "Empty PrinterIP
// disables actual network dispatch" contract. Without the guard,
// net.JoinHostPort("", "9100") dials localhost:9100 at request time.
func Test_TCPPrinter_EmptyHostNoop(t *testing.T) {
	p := tcpprint.New("", "9100", 500*time.Millisecond)
	if err := p.SendZPL(context.Background(), []byte("^XA^XZ")); err != nil {
		t.Fatalf("SendZPL with empty host should be no-op, got: %v", err)
	}
}
