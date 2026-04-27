// Package zpl contains ZPL template functions for printable labels.
// Templates are byte-equivalent ports of src/utils/zpl/*.ts (deleted in 0b.14).
// JSON tags match the TS camelCase field names so Render can unmarshal
// payloads sent from the frontend directly.
package zpl

// ReceivingData mirrors ReceivingLabelData from src/utils/zpl/types.ts.
type ReceivingData struct {
	ProductName string  `json:"productName"`
	SKU         string  `json:"sku"`
	UPC         string  `json:"upc"`
	LotNumber   *string `json:"lotNumber"`
	ExpiryDate  *string `json:"expiryDate"`
	Quantity    int     `json:"quantity"`
	PONumber    string  `json:"poNumber"`
}

// PickData mirrors PickLabelData from src/utils/zpl/types.ts.
type PickData struct {
	OrderNumber   string   `json:"orderNumber"`
	CustomerName  string   `json:"customerName"`
	ProductName   string   `json:"productName"`
	SKU           string   `json:"sku"`
	UPC           string   `json:"upc"`
	LotNumber     *string  `json:"lotNumber"`
	SerialNumbers []string `json:"serialNumbers"`
	Quantity      int      `json:"quantity"`
	LocationCode  string   `json:"locationCode"`
}

// LocationData is for new location labels (no TS source).
type LocationData struct {
	Code string `json:"code"`
}

// ToteData is for new tote labels (no TS source).
type ToteData struct {
	Code string `json:"code"`
}

// ProductData is for new product (Layer B item-identity) labels.
// Per design doc 2026-04-24 §3.1: one label_catalog row per product SKU,
// entity_ref = product_id, applied to each inbound case at receive.
type ProductData struct {
	ProductName string  `json:"productName"`
	SKU         string  `json:"sku"`
	UPC         string  `json:"upc"`
	LotNumber   *string `json:"lotNumber"`
}
