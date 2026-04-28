// business/domain/paperwork/paperworkbus/pdf/pdf_helpers_test.go
//
// Test helpers for the pdf package. Per plan VFY-2 fallback, these
// rely on raw byte scanning of uncompressed gofpdf output (renderers
// call SetCompression(false)) instead of pulling pdfcpu/pdfcpu — that
// dep would have forced a Go 1.23→1.25 directive bump that breaks the
// dockerfile.ichor golang:1.23 pin.
package pdf_test

import (
	"bytes"
	"testing"
)

// pdfPageCount returns the number of /Page objects in the PDF stream.
// Counts "/Type /Page\n" — the trailing newline distinguishes /Page
// objects from the /Pages parent (which has trailing 's').
func pdfPageCount(t *testing.T, b []byte) (int, error) {
	t.Helper()
	return bytes.Count(b, []byte("/Type /Page\n")), nil
}

// pdfContainsText reports whether the rendered PDF contains the given
// human-readable text in its uncompressed content stream. Renderers
// must call SetCompression(false) for this to work.
func pdfContainsText(t *testing.T, b []byte, needle string) bool {
	t.Helper()
	return bytes.Contains(b, []byte(needle))
}
