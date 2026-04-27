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

	y := 60
	lines := []string{
		"^XA",
		fmt.Sprintf("^FO40,%d^A0N,60,60^FD%s^FS", y, name),
	}

	y += 90
	lines = append(lines, fmt.Sprintf("^FO40,%d^A0N,40,40^FDSKU: %s^FS", y, d.SKU))

	y += 80
	lines = append(lines, fmt.Sprintf("^FO40,%d^BY3^BCN,120,Y,N,N^FD%s^FS", y, d.UPC))

	y += 200
	if d.LotNumber != nil {
		lines = append(lines, fmt.Sprintf("^FO40,%d^A0N,40,40^FDLOT: %s^FS", y, *d.LotNumber))
	}

	lines = append(lines, "^XZ")
	return strings.Join(lines, "\n")
}
