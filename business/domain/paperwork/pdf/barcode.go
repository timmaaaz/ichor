// Package pdf renders paperwork PDFs (pick sheets, receive covers, transfer
// sheets) for floor-worker workflows. Each PDF carries a scannable Code128
// task code at the top — scanning the sheet routes the worker to the
// corresponding composable. Phase 0g.B3 introduces this package; Phase 0g.B2
// scaffolded the surrounding bus/app/api layers.
package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

// Code128PNG renders content as a Code128 barcode PNG, scaled for legible
// thermal/ink rendering. Returns the encoded PNG bytes.
//
// Determinism: same input → same bytes (boombuler is deterministic; png.Encode
// is deterministic at default compression).
func Code128PNG(content string) ([]byte, error) {
	if content == "" {
		return nil, errors.New("pdf: Code128PNG content must not be empty")
	}

	bc, err := code128.Encode(content)
	if err != nil {
		return nil, fmt.Errorf("pdf: code128 encode %q: %w", content, err)
	}

	// 600×100 px — sized for crisp rendering at 200 DPI in a ~3 in (76 mm)
	// barcode cell on US-letter paper. Width chosen to leave whitespace
	// margins in the final PDF cell after fpdf scales the embedded image.
	scaled, err := barcode.Scale(bc, 600, 100)
	if err != nil {
		return nil, fmt.Errorf("pdf: code128 scale: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, fmt.Errorf("pdf: png encode: %w", err)
	}
	return buf.Bytes(), nil
}
