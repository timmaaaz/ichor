// Package zpl contains ZPL template functions for printable labels.
// Templates are pure string-builders; see render.go for the dispatcher.
package zpl

// LocationData is for location labels.
type LocationData struct {
	Code string `json:"code"`
}

// ToteData is for container (tote/bin/carton) labels.
type ToteData struct {
	Code string `json:"code"`
}

// ProductData is for product (Layer B item-identity) labels.
// Per design doc 2026-04-24 §3.1: one label_catalog row per product SKU,
// entity_ref = product_id, applied to each inbound case at receive.
type ProductData struct {
	ProductName string  `json:"productName"`
	SKU         string  `json:"sku"`
	UPC         string  `json:"upc"`
	LotNumber   *string `json:"lotNumber"`
}
