// app/sdk/auth/context.go
package auth

import "context"

// contextKey is a private type for context keys used by the auth package
type contextKey int

// Define keys for the context
const (
	tableInfoKey contextKey = iota
)

// WithTableInfo adds table information to the context
func WithTableInfo(ctx context.Context, tableInfo *TableInfo) context.Context {
	return context.WithValue(ctx, tableInfoKey, tableInfo)
}

// GetTableInfo extracts table information from the context
func GetTableInfo(ctx context.Context) (*TableInfo, bool) {
	tableInfo, ok := ctx.Value(tableInfoKey).(*TableInfo)
	return tableInfo, ok
}
