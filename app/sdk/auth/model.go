package auth

// TableInfo represents the structure and metadata of a database table
type TableInfo struct {
	Name   string // Table name
	Action Action // "create", "read", "update", "delete"
}
