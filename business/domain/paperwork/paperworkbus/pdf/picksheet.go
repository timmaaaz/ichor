package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"

	"github.com/go-pdf/fpdf"
)

// PickSheet renders a letter-size pick sheet PDF. Header carries a Code128
// task code (data.TaskCode) plus human-readable label; body carries a
// line-items table.
//
// Errors:
//   - empty TaskCode → returned as a regular error (caller should never
//     pass empty; bus enforces non-empty before calling).
//   - go-pdf/fpdf failures bubble up wrapped.
func PickSheet(data PickSheetData) ([]byte, error) {
	if data.TaskCode == "" {
		return nil, errors.New("pdf: PickSheet requires non-empty TaskCode")
	}

	pdfDoc := fpdf.New("P", "mm", "Letter", "")
	// Disable content-stream compression so test helpers can do raw
	// bytes.Contains scans on rendered text. Production paperwork is not
	// large enough for compression to matter (single-page, < 50 KB).
	pdfDoc.SetCompression(false)
	pdfDoc.AddPage()

	if err := drawHeader(pdfDoc, "Pick Sheet", data.TaskCode); err != nil {
		return nil, fmt.Errorf("pdf: pick sheet header: %w", err)
	}

	pdfDoc.SetFont("Helvetica", "", 10)
	pdfDoc.Cell(0, 5, fmt.Sprintf("Order: %s", data.OrderNumber))
	pdfDoc.Ln(5)
	pdfDoc.Cell(0, 5, fmt.Sprintf("Customer: %s", data.CustomerName))
	pdfDoc.Ln(5)
	if data.Zone != "" {
		pdfDoc.Cell(0, 5, fmt.Sprintf("Zone: %s", data.Zone))
		pdfDoc.Ln(5)
	}
	pdfDoc.Ln(5)

	// Column widths sum to 180mm (letter width 215.9mm minus 2×10mm
	// margins = 195.9mm, with 15.9mm right slack to keep cell text
	// from running into the right margin).
	pdfDoc.SetFont("Helvetica", "B", 9)
	pdfDoc.Cell(35, 6, "Location")
	pdfDoc.Cell(40, 6, "SKU")
	pdfDoc.Cell(85, 6, "Product")
	pdfDoc.Cell(20, 6, "Qty")
	pdfDoc.Ln(6)

	pdfDoc.SetFont("Helvetica", "", 9)
	for _, line := range data.Lines {
		pdfDoc.Cell(35, 6, line.LocationCode)
		pdfDoc.Cell(40, 6, line.SKU)
		pdfDoc.Cell(85, 6, line.ProductName)
		pdfDoc.Cell(20, 6, fmt.Sprintf("%d", line.Quantity))
		pdfDoc.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdfDoc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output pick sheet: %w", err)
	}
	return buf.Bytes(), nil
}

// drawHeader renders a Code128 image (top-left) plus large plain-text task
// code label below. Shared by all three sheet renderers (PickSheet,
// ReceiveCover, TransferSheet — added in subsequent commits).
func drawHeader(pdfDoc *fpdf.Fpdf, title, taskCode string) error {
	pngBytes, err := Code128PNG(taskCode)
	if err != nil {
		return fmt.Errorf("code128 png: %w", err)
	}

	// Sanity-check the PNG before handing to fpdf.RegisterImageOptionsReader (fpdf
	// otherwise panics on malformed image streams).
	if _, err := png.Decode(bytes.NewReader(pngBytes)); err != nil {
		return fmt.Errorf("decode generated png: %w", err)
	}

	imgName := "code128-" + taskCode
	pdfDoc.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(pngBytes))
	pdfDoc.ImageOptions(imgName, 10, 10, 80, 13, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	pdfDoc.SetXY(10, 26)
	pdfDoc.SetFont("Helvetica", "B", 14)
	pdfDoc.Cell(0, 6, taskCode)
	pdfDoc.Ln(8)

	pdfDoc.SetFont("Helvetica", "B", 16)
	pdfDoc.Cell(0, 8, title)
	pdfDoc.Ln(10)

	return nil
}
