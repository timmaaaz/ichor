package zpl

import (
	"fmt"
	"strings"
)

// Product renders a 4"×6" @ 203 DPI ZPL product (Layer B item-identity) label.
// Applied to each inbound case at receive; persists with the case through put-away,
// storage, pick, and ship per design doc 2026-04-24 §3.1.
//
// Layout (203 DPI: 1" = 203 dots; label is 812w × 1218h):
//   Header (product name, truncated at 30 chars) — top
//   SKU line — below header
//   UPC Code128 barcode — middle
//   LOT line (optional) — below barcode
func Product(d ProductData) string {
	name := d.ProductName
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	var b strings.Builder
	b.WriteString("^XA\n")

	y := 60
	b.WriteString(fmt.Sprintf("^FO40,%d^A0N,60,60^FD%s^FS\n", y, name))

	y += 90
	b.WriteString(fmt.Sprintf("^FO40,%d^A0N,40,40^FDSKU: %s^FS\n", y, d.SKU))

	y += 80
	b.WriteString(fmt.Sprintf("^FO40,%d^BY3^BCN,120,Y,N,N^FD%s^FS\n", y, d.UPC))

	y += 200
	if d.LotNumber != nil {
		b.WriteString(fmt.Sprintf("^FO40,%d^A0N,40,40^FDLOT: %s^FS\n", y, *d.LotNumber))
	}

	b.WriteString("^XZ\n")
	return b.String()
}
