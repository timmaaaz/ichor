package pdf

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// TransferSheet renders a letter-size transfer sheet PDF. Header carries a
// Code128 task code (XFER-*) and the body lists transfer lines.
// Reuses drawHeader from picksheet.go.
func TransferSheet(data TransferSheetData) ([]byte, error) {
	if data.TaskCode == "" {
		return nil, errors.New("pdf: TransferSheet requires non-empty TaskCode")
	}

	pdfDoc := fpdf.New("P", "mm", "Letter", "")
	// Disable content-stream compression so test helpers can do raw
	// bytes.Contains scans on rendered text. Production paperwork is not
	// large enough for compression to matter (single-page, < 50 KB).
	pdfDoc.SetCompression(false)
	pdfDoc.AddPage()

	if err := drawHeader(pdfDoc, "Transfer Sheet", data.TaskCode); err != nil {
		return nil, fmt.Errorf("pdf: transfer sheet header: %w", err)
	}

	pdfDoc.SetFont("Helvetica", "", 10)
	pdfDoc.Cell(0, 5, fmt.Sprintf("Transfer: %s", data.TransferNum))
	pdfDoc.Ln(5)
	pdfDoc.Cell(0, 5, fmt.Sprintf("From: %s   To: %s", data.SourceWH, data.DestinationWH))
	pdfDoc.Ln(5)
	pdfDoc.Ln(5)

	// Column widths sum to 190mm (letter width 215.9mm minus 2×10mm
	// margins = 195.9mm, with 5.9mm right slack to keep cell text
	// from running into the right margin).
	pdfDoc.SetFont("Helvetica", "B", 9)
	pdfDoc.Cell(40, 6, "SKU")
	pdfDoc.Cell(70, 6, "Product")
	pdfDoc.Cell(30, 6, "From")
	pdfDoc.Cell(30, 6, "To")
	pdfDoc.Cell(20, 6, "Qty")
	pdfDoc.Ln(6)

	pdfDoc.SetFont("Helvetica", "", 9)
	for _, line := range data.Lines {
		pdfDoc.Cell(40, 6, line.SKU)
		pdfDoc.Cell(70, 6, line.ProductName)
		pdfDoc.Cell(30, 6, line.FromLocation)
		pdfDoc.Cell(30, 6, line.ToLocation)
		pdfDoc.Cell(20, 6, fmt.Sprintf("%d", line.Quantity))
		pdfDoc.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdfDoc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output transfer sheet: %w", err)
	}
	return buf.Bytes(), nil
}
