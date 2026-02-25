// Package query provides support for query paging.
package query

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// Result is the data model used when returning a query result.
type Result[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rows_per_page"`
}

// NewResult constructs a result value to return query results.
func NewResult[T any](items []T, total int, page page.Page) Result[T] {
	return Result[T]{
		Items:       items,
		Total:       total,
		Page:        page.Number(),
		RowsPerPage: page.RowsPerPage(),
	}
}

// Encode implements the encoder interface.
func (r Result[T]) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// ParseIDs converts a slice of string UUIDs into a slice of uuid.UUID values.
// Used by batch query endpoints to parse IDs from request bodies.
func ParseIDs(ids []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("parse id[%d]: %w", i, err)
		}
		uuids[i] = uid
	}
	return uuids, nil
}
