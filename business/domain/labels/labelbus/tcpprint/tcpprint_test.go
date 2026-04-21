package tcpprint_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/tcpprint"
)

func Test_TCPPrinter_SendZPL(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		data, _ := io.ReadAll(conn)
		received <- string(data)
	}()

	p := tcpprint.New(ln.Addr().String(), 2*time.Second)

	payload := "^XA\n^FDHELLO^FS\n^XZ\n"
	if err := p.SendZPL(context.Background(), []byte(payload)); err != nil {
		t.Fatalf("SendZPL: %v", err)
	}

	select {
	case got := <-received:
		if got != payload {
			t.Fatalf("bytes mismatch: want %q got %q", payload, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for bytes")
	}
}

func Test_TCPPrinter_DialFailure(t *testing.T) {
	// 127.0.0.1:1 is virtually guaranteed to refuse; timeout is hard cap.
	p := tcpprint.New("127.0.0.1:1", 500*time.Millisecond)
	err := p.SendZPL(context.Background(), []byte("^XA^XZ"))
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
}
