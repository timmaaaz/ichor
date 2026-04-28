// business/domain/paperwork/paperworkbus/pdf/picksheet_test.go
package pdf_test

import (
	"bytes"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus/pdf"
)

func TestPickSheet_StructuralInvariants(t *testing.T) {
	t.Parallel()

	data := pdf.PickSheetData{
		TaskCode:     "SO-12345",
		OrderNumber:  "12345",
		CustomerName: "ACME Co",
		Zone:         "",
		Lines: []pdf.PickSheetLine{
			{LocationCode: "STG-A01", SKU: "SKU-0001", ProductName: "Widget", Quantity: 3},
			{LocationCode: "STG-B02", SKU: "SKU-0042", ProductName: "Gadget", Quantity: 1},
		},
	}

	got, err := pdf.PickSheet(data)
	if err != nil {
		t.Fatalf("PickSheet: %v", err)
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

	// Plain-text taskCode + customer + line-item SKU appear in the PDF
	// content stream. gofpdf streams text uncompressed by default, so
	// the helper does a raw bytes.Contains scan rather than structured
	// extraction. Image bytes (Code128) are not scanned by this helper —
	// image presence is asserted via image-count if needed; structural
	// invariants here cover the human-readable text only.
	wantStrings := []string{"SO-12345", "ACME Co", "SKU-0042"}
	for _, want := range wantStrings {
		if !pdfContainsText(t, got, want) {
			t.Errorf("PDF missing expected text %q", want)
		}
	}
}

func TestPickSheet_EmptyTaskCode(t *testing.T) {
	t.Parallel()

	if _, err := pdf.PickSheet(pdf.PickSheetData{TaskCode: ""}); err == nil {
		t.Fatal("PickSheet with empty TaskCode returned nil error; want failure")
	}
}
