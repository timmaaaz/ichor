package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterAllTools registers every tool on the server (context = "all").
func RegisterAllTools(s *mcp.Server, c *client.Client) {
	RegisterDiscoveryTools(s, c)
	RegisterUIReadTools(s, c)
	RegisterWorkflowReadTools(s, c)
	RegisterSearchTools(s, c)
	RegisterWorkflowWriteTools(s, c)
	RegisterUIWriteTools(s, c)
	RegisterValidationTools(s, c)
	RegisterAnalysisTools(s, c)
}

// RegisterToolsForContext registers only the tools that belong to the given
// context mode: "all", "workflow", or "tables".
func RegisterToolsForContext(s *mcp.Server, c *client.Client, contextMode string) {
	switch contextMode {
	case "workflow":
		RegisterWorkflowDiscoveryTools(s, c)
		RegisterWorkflowReadTools(s, c)
		RegisterSearchTools(s, c)
		RegisterWorkflowWriteTools(s, c)
		RegisterAnalysisTools(s, c)
	case "tables":
		RegisterTablesDiscoveryTools(s, c)
		RegisterUIReadTools(s, c)
		RegisterSearchTools(s, c)
		RegisterUIWriteTools(s, c)
		RegisterValidationTools(s, c)
	default: // "all"
		RegisterAllTools(s, c)
	}
}
