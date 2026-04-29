package pdf_test

import (
	"bytes"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus/pdf"
)

func TestTransferSheet_StructuralInvariants(t *testing.T) {
	t.Parallel()

	data := pdf.TransferSheetData{
		TaskCode:      "XFER-251015-0001",
		TransferNum:   "XFER-251015-0001",
		SourceWH:      "WH-A",
		DestinationWH: "WH-B",
		Lines: []pdf.TransferSheetLine{
			{SKU: "SKU-0001", ProductName: "Widget", FromLocation: "STG-A01", ToLocation: "STG-C03", Quantity: 5},
		},
	}

	got, err := pdf.TransferSheet(data)
	if err != nil {
		t.Fatalf("TransferSheet: %v", err)
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

	for _, want := range []string{"XFER-251015-0001", "WH-A", "WH-B", "SKU-0001"} {
		if !pdfContainsText(t, got, want) {
			t.Errorf("PDF missing expected text %q", want)
		}
	}
}

func TestTransferSheet_EmptyTaskCode(t *testing.T) {
	t.Parallel()
	if _, err := pdf.TransferSheet(pdf.TransferSheetData{TaskCode: ""}); err == nil {
		t.Fatal("TransferSheet with empty TaskCode returned nil error; want failure")
	}
}
