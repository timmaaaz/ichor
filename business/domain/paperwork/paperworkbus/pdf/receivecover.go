package pdf

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// ReceiveCover renders a letter-size receive-cover PDF. Header carries a
// Code128 task code (PO-*) and the body lists expected lines for the PO.
// Reuses drawHeader from picksheet.go.
//
// Errors:
//   - empty TaskCode → returned as a regular error (caller should never
//     pass empty; bus enforces non-empty before calling).
//   - go-pdf/fpdf failures bubble up wrapped.
func ReceiveCover(data ReceiveCoverData) ([]byte, error) {
	if data.TaskCode == "" {
		return nil, errors.New("pdf: ReceiveCover requires non-empty TaskCode")
	}

	pdfDoc := fpdf.New("P", "mm", "Letter", "")
	// Disable content-stream compression so test helpers can do raw
	// bytes.Contains scans on rendered text. Production paperwork is not
	// large enough for compression to matter (single-page, < 50 KB).
	pdfDoc.SetCompression(false)
	pdfDoc.AddPage()

	if err := drawHeader(pdfDoc, "Receive Cover", data.TaskCode); err != nil {
		return nil, fmt.Errorf("pdf: receive cover header: %w", err)
	}

	pdfDoc.SetFont("Helvetica", "", 10)
	pdfDoc.Cell(0, 5, fmt.Sprintf("PO Number: %s", data.PONumber))
	pdfDoc.Ln(5)
	pdfDoc.Cell(0, 5, fmt.Sprintf("Vendor: %s", data.VendorName))
	pdfDoc.Ln(5)
	if data.ExpectedAt != "" {
		pdfDoc.Cell(0, 5, fmt.Sprintf("Expected: %s", data.ExpectedAt))
		pdfDoc.Ln(5)
	}
	pdfDoc.Ln(5)

	// Column widths sum to 180mm (letter width 215.9mm minus 2×10mm
	// margins = 195.9mm, with 15.9mm right slack to keep cell text
	// from running into the right margin).
	pdfDoc.SetFont("Helvetica", "B", 9)
	pdfDoc.Cell(40, 6, "SKU")
	pdfDoc.Cell(85, 6, "Product")
	pdfDoc.Cell(35, 6, "UPC")
	pdfDoc.Cell(20, 6, "Expected")
	pdfDoc.Ln(6)

	pdfDoc.SetFont("Helvetica", "", 9)
	for _, line := range data.Lines {
		pdfDoc.Cell(40, 6, line.SKU)
		pdfDoc.Cell(85, 6, line.ProductName)
		pdfDoc.Cell(35, 6, line.UPC)
		pdfDoc.Cell(20, 6, fmt.Sprintf("%d", line.Expected))
		pdfDoc.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdfDoc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output receive cover: %w", err)
	}
	return buf.Bytes(), nil
}
