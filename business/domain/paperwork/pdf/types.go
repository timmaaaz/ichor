package pdf

// PickSheetData holds the rendering inputs for a pick sheet PDF.
//
// The renderer is decoupled from sibling-bus types: the bus layer queries
// ordersbus, transforms the result into PickSheetData, and calls PickSheet.
// This keeps the pdf package free of cross-domain imports.
type PickSheetData struct {
	TaskCode     string // Code128 content + plain-text label (e.g. "SO-12345")
	OrderNumber  string // human-readable order id (e.g. "12345")
	CustomerName string
	Zone         string // optional; empty = all zones
	Lines        []PickSheetLine
}

// PickSheetLine is one row in the pick sheet table.
type PickSheetLine struct {
	LocationCode string
	SKU          string
	ProductName  string
	Quantity     int
}

// ReceiveCoverData holds the rendering inputs for a receive-cover PDF.
type ReceiveCoverData struct {
	TaskCode   string // e.g. "PO-1"
	PONumber   string
	VendorName string
	ExpectedAt string // ISO date or empty
	Lines      []ReceiveCoverLine
}

// ReceiveCoverLine is one row in the receive-cover expected-lines table.
type ReceiveCoverLine struct {
	SKU         string
	ProductName string
	UPC         string
	Expected    int
}

// TransferSheetData holds the rendering inputs for a transfer sheet PDF.
type TransferSheetData struct {
	TaskCode      string // e.g. "XFER-251015-0001"
	TransferNum   string
	SourceWH      string
	DestinationWH string
	Lines         []TransferSheetLine
}

// TransferSheetLine is one row in the transfer-sheet table.
type TransferSheetLine struct {
	SKU          string
	ProductName  string
	FromLocation string
	ToLocation   string
	Quantity     int
}
