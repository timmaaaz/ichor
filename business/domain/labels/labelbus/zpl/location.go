package zpl

import (
	"fmt"
	"strings"
)

// Location renders a 2"×1" @ 203 DPI ZPL location label:
// human-readable code + Code128 barcode of the same code.
// Matches the Phase 0 smoke pattern (NOTES.md 2026-04-14).
func Location(d LocationData) string {
	var b strings.Builder
	b.WriteString("^XA\n")
	b.WriteString(fmt.Sprintf("^FO50,50^A0N,40,40^FD%s^FS\n", d.Code))
	b.WriteString(fmt.Sprintf("^FO50,100^BY2^BCN,80,Y,N,N^FD%s^FS\n", d.Code))
	b.WriteString("^XZ\n")
	return b.String()
}
