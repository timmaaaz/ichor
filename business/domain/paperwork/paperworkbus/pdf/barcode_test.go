// Package pdf_test verifies paperwork PDF helpers in isolation.
package pdf_test

import (
	"bytes"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus/pdf"
)

func TestCode128PNG_Determinism(t *testing.T) {
	t.Parallel()

	a, err := pdf.Code128PNG("SO-12345")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := pdf.Code128PNG("SO-12345")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("non-deterministic Code128 PNG output: same input produced different bytes")
	}
}

func TestCode128PNG_PNGSignature(t *testing.T) {
	t.Parallel()

	got, err := pdf.Code128PNG("PO-1")
	if err != nil {
		t.Fatalf("Code128PNG: %v", err)
	}
	want := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(got) < len(want) || !bytes.Equal(got[:len(want)], want) {
		t.Fatalf("output is not a PNG: first 8 bytes = %x", got[:min(8, len(got))])
	}
}

func TestCode128PNG_EmptyContent(t *testing.T) {
	t.Parallel()

	if _, err := pdf.Code128PNG(""); err == nil {
		t.Fatal("Code128PNG(\"\") returned nil error; want failure")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
