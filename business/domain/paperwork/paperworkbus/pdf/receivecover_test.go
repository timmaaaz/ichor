// business/domain/paperwork/paperworkbus/pdf/receivecover_test.go
package pdf_test

import (
	"bytes"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus/pdf"
)

func TestReceiveCover_StructuralInvariants(t *testing.T) {
	t.Parallel()

	data := pdf.ReceiveCoverData{
		TaskCode:   "PO-1",
		PONumber:   "1",
		VendorName: "Acme Supply",
		ExpectedAt: "2026-04-30",
		Lines: []pdf.ReceiveCoverLine{
			{SKU: "SKU-0001", ProductName: "Widget", UPC: "012345678905", Expected: 24},
			{SKU: "SKU-0042", ProductName: "Gadget", UPC: "012345678912", Expected: 12},
		},
	}

	got, err := pdf.ReceiveCover(data)
	if err != nil {
		t.Fatalf("ReceiveCover: %v", err)
	}

	if !bytes.HasPrefix(got, []byte("%PDF-")) {
		t.Fatalf("output is not a PDF: first 5 bytes = %q", got[:5])
	}

	pageCount, err := pdfPageCount(t, got)
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("page count: got %d, want 1", pageCount)
	}

	for _, want := range []string{"PO-1", "Acme Supply", "SKU-0042", "012345678912"} {
		if !pdfContainsText(t, got, want) {
			t.Errorf("PDF missing expected text %q", want)
		}
	}
}

func TestReceiveCover_EmptyTaskCode(t *testing.T) {
	t.Parallel()
	if _, err := pdf.ReceiveCover(pdf.ReceiveCoverData{TaskCode: ""}); err == nil {
		t.Fatal("ReceiveCover with empty TaskCode returned nil error; want failure")
	}
}
