package zpl

import (
	"fmt"
	"strings"
)

// Receiving renders a 4"×2" @ 203 DPI ZPL receiving label.
// Byte-equivalent port of src/utils/zpl/receivingLabel.ts.
func Receiving(d ReceivingData) string {
	name := d.ProductName
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	y := 30
	lines := []string{
		"^XA",
		fmt.Sprintf("^FO30,%d^A0N,36,36^FD%s^FS", y, name),
	}

	y += 45
	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,24,24^FDSKU: %s^FS", y, d.SKU))

	y += 40
	lines = append(lines, fmt.Sprintf("^FO30,%d^BCN,60,Y,N,N^FD%s^FS", y, d.UPC))

	y += 95
	if d.LotNumber != nil {
		var lotLine string
		if d.ExpiryDate != nil {
			lotLine = fmt.Sprintf("LOT: %s  EXP: %s", *d.LotNumber, *d.ExpiryDate)
		} else {
			lotLine = fmt.Sprintf("LOT: %s", *d.LotNumber)
		}
		lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,24,24^FD%s^FS", y, lotLine))
		y += 35
	}

	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,24,24^FDQTY: %d  PO: %s^FS", y, d.Quantity, d.PONumber))

	lines = append(lines, "^XZ")
	return strings.Join(lines, "\n")
}
