// Package tcpprint writes ZPL bytes directly to a Zebra-compatible TCP
// printer on port 9100 (PDL-datastream). One-shot connection per print;
// no pooling (printers often reset the connection after each job).
package tcpprint

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Printer sends ZPL bytes to a network-attached thermal printer.
type Printer struct {
	host    string
	port    string
	timeout time.Duration
}

// New constructs a Printer. `timeout` caps both dial and write durations.
func New(host, port string, timeout time.Duration) *Printer {
	return &Printer{host: host, port: port, timeout: timeout}
}

// SendZPL opens a TCP connection, writes the ZPL bytes, and closes.
// A single retry is attempted on dial failure after a 250ms backoff,
// since printers occasionally refuse the first connection after idle.
//
// An empty host is a no-op: this honours the mux.Config contract that
// "Empty PrinterIP disables actual network dispatch". Without the
// guard, net.JoinHostPort("", port) yields ":9100" and the printer
// dials localhost at request time, which is both unsafe and
// surprising in environments that intentionally leave PrinterIP unset.
func (p *Printer) SendZPL(ctx context.Context, zpl []byte) error {
	if p.host == "" {
		return nil
	}
	d := net.Dialer{Timeout: p.timeout}
	addr := net.JoinHostPort(p.host, p.port)

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
		conn, err = d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("dial %s: %w", addr, err)
		}
	}
	defer conn.Close()

	deadline := time.Now().Add(p.timeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetWriteDeadline(deadline)

	n, err := conn.Write(zpl)
	if err != nil {
		return fmt.Errorf("write zpl: %w", err)
	}
	if n != len(zpl) {
		return fmt.Errorf("write zpl: short write: wrote %d of %d bytes", n, len(zpl))
	}
	return nil
}
