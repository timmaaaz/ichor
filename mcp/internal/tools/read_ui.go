package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterUIReadTools registers all UI config read tools. It is an aggregator
// that delegates to the granular registration functions for pages, page actions,
// and content blocks (tables, forms).
func RegisterUIReadTools(s *mcp.Server, c *client.Client) {
	RegisterPageReadTools(s, c)
	RegisterPageActionReadTools(s, c)
	RegisterContentBlockReadTools(s, c)
}
