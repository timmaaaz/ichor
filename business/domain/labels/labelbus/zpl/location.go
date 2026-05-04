package zpl

import (
	"fmt"
	"strings"
)

// Location renders a 4"×6" @ 203 DPI ZPL location label:
// large human-readable code + Code128 barcode of the same code.
// Sized for Zebra GK420t with 4×6 thermal-transfer media.
//
// Layout (203 DPI: 1" = 203 dots; label is 812w × 1218h):
//   Code text (150-dot font) — top
//   Code128 barcode (BY4, 250-dot height, with human-readable) — upper third
func Location(d LocationData) string {
	var b strings.Builder
	b.WriteString("^XA\n")
	b.WriteString(fmt.Sprintf("^FO40,80^A0N,150,150^FD%s^FS\n", d.Code))
	b.WriteString(fmt.Sprintf("^FO40,300^BY4^BCN,250,Y,N,N^FD%s^FS\n", d.Code))
	b.WriteString("^XZ\n")
	return b.String()
}
