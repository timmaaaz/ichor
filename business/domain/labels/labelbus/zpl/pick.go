package zpl

import (
	"fmt"
	"strings"
)

// Pick renders a 4"×2" @ 203 DPI ZPL pick label.
// Byte-equivalent port of src/utils/zpl/pickLabel.ts.
func Pick(d PickData) string {
	name := d.ProductName
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	customer := d.CustomerName
	if len(customer) > 25 {
		customer = customer[:22] + "..."
	}

	y := 20
	lines := []string{
		"^XA",
		fmt.Sprintf("^FO30,%d^A0N,28,28^FDOrder: %s^FS", y, d.OrderNumber),
	}

	y += 32
	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,22,22^FD%s^FS", y, customer))

	y += 30
	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,28,28^FD%s^FS", y, name))

	y += 35
	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,22,22^FDSKU: %s^FS", y, d.SKU))

	y += 32
	lines = append(lines, fmt.Sprintf("^FO30,%d^BCN,50,Y,N,N^FD%s^FS", y, d.UPC))

	y += 80
	if d.LotNumber != nil {
		lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,22,22^FDLOT: %s  QTY: %d^FS", y, *d.LotNumber, d.Quantity))
	} else {
		lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,22,22^FDQTY: %d^FS", y, d.Quantity))
	}

	y += 30
	if len(d.SerialNumbers) > 0 {
		serials := strings.Join(d.SerialNumbers, ", ")
		truncated := serials
		if len(truncated) > 50 {
			truncated = truncated[:47] + "..."
		}
		lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,20,20^FDS/N: %s^FS", y, truncated))
		y += 28
	}

	lines = append(lines, fmt.Sprintf("^FO30,%d^A0N,22,22^FDFrom: %s^FS", y, d.LocationCode))

	lines = append(lines, "^XZ")
	return strings.Join(lines, "\n")
}
