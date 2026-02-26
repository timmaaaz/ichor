package scanapp

import "encoding/json"

// ScanResult is the top-level response for a barcode scan.
type ScanResult struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (s ScanResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}

// ProductScanResult is returned when the barcode matches a product UPC code.
type ProductScanResult struct {
	ProductID    string              `json:"product_id"`
	Name         string              `json:"name"`
	SKU          string              `json:"sku"`
	TrackingType string              `json:"tracking_type"`
	StockSummary []StockAtLocation   `json:"stock_summary"`
}

// StockAtLocation summarises inventory quantities at one location.
type StockAtLocation struct {
	LocationID       string `json:"location_id"`
	LocationCode     string `json:"location_code"`
	Quantity         int    `json:"quantity"`
	ReservedQuantity int    `json:"reserved_quantity"`
}

// LocationScanResult is returned when the barcode matches a location code exactly.
type LocationScanResult struct {
	LocationID   string           `json:"location_id"`
	LocationCode string           `json:"location_code"`
	Aisle        string           `json:"aisle"`
	Rack         string           `json:"rack"`
	Shelf        string           `json:"shelf"`
	Bin          string           `json:"bin"`
	Items        []ItemAtLocation `json:"items"`
}

// ItemAtLocation describes a product stocked at a scanned location.
type ItemAtLocation struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	ProductSKU   string `json:"product_sku"`
	TrackingType string `json:"tracking_type"`
	Quantity     int    `json:"quantity"`
}

// LotScanResult is returned when the barcode matches a lot number.
type LotScanResult struct {
	LotID         string             `json:"lot_id"`
	LotNumber     string             `json:"lot_number"`
	ProductID     string             `json:"product_id"`
	ProductName   string             `json:"product_name"`
	ProductSKU    string             `json:"product_sku"`
	QualityStatus string             `json:"quality_status"`
	Quantity      int                `json:"quantity"`
	Locations     []LotLocationEntry `json:"locations"`
}

// LotLocationEntry describes a location where units of the lot are stored.
type LotLocationEntry struct {
	LocationID   string `json:"location_id"`
	LocationCode string `json:"location_code"`
	Aisle        string `json:"aisle"`
	Rack         string `json:"rack"`
	Shelf        string `json:"shelf"`
	Bin          string `json:"bin"`
	Quantity     int    `json:"quantity"`
}

// SerialScanResult is returned when the barcode matches a serial number.
type SerialScanResult struct {
	SerialID      string `json:"serial_id"`
	SerialNumber  string `json:"serial_number"`
	ProductID     string `json:"product_id"`
	LotID         string `json:"lot_id"`
	Status        string `json:"status"`
	LocationID    string `json:"location_id"`
	LocationCode  string `json:"location_code"`
	Aisle         string `json:"aisle"`
	Rack          string `json:"rack"`
	Shelf         string `json:"shelf"`
	Bin           string `json:"bin"`
	WarehouseName string `json:"warehouse_name"`
	ZoneName      string `json:"zone_name"`
}
