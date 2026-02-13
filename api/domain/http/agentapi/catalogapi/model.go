package catalogapi

import (
	"encoding/json"
)

// Endpoints describes the CRUD and query URLs for a config surface.
type Endpoints struct {
	List   string `json:"list,omitempty"`
	Get    string `json:"get,omitempty"`
	Create string `json:"create,omitempty"`
	Update string `json:"update,omitempty"`
	Delete string `json:"delete,omitempty"`
}

// ConfigSurface describes a single configurable area of the system.
type ConfigSurface struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Endpoints    Endpoints `json:"endpoints"`
	DiscoveryURL string   `json:"discovery_url,omitempty"`
	Constraints  []string `json:"constraints,omitempty"`
}

// Catalog is a slice of ConfigSurface for API responses.
type Catalog []ConfigSurface

// Encode implements web.Encoder.
func (c Catalog) Encode() ([]byte, string, error) {
	data, err := json.Marshal(c)
	return data, "application/json", err
}
