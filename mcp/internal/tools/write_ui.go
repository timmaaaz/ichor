package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterUIWriteTools registers all UI config mutation tools. It is an
// aggregator that delegates to the granular registration functions for pages
// and content blocks (tables, forms).
func RegisterUIWriteTools(s *mcp.Server, c *client.Client) {
	RegisterPageWriteTools(s, c)
	RegisterContentBlockWriteTools(s, c)
}
