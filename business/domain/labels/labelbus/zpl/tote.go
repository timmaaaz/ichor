package zpl

import (
	"fmt"
	"strings"
)

// Tote renders a 2"×1" @ 203 DPI ZPL tote label.
// Shares layout with Location at Phase 0b; diverges in Phase 1+ when
// totes gain lot-expiry/icon fields. Kept as a separate function so
// call sites read intention rather than implementation.
func Tote(d ToteData) string {
	var b strings.Builder
	b.WriteString("^XA\n")
	b.WriteString(fmt.Sprintf("^FO50,50^A0N,40,40^FD%s^FS\n", d.Code))
	b.WriteString(fmt.Sprintf("^FO50,100^BY2^BCN,80,Y,N,N^FD%s^FS\n", d.Code))
	b.WriteString("^XZ\n")
	return b.String()
}
