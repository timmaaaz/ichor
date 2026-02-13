// Command ichor-mcp runs an MCP (Model Context Protocol) server that wraps
// the Ichor REST API, providing discovery, read, and search tools for LLM agents.
//
// Usage:
//
//	ichor-mcp --api-url http://localhost:8080 --token $TOKEN
//
// The server communicates over stdio using JSON-RPC, compatible with Claude Desktop,
// Ollama, and other MCP-capable clients.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
	"github.com/timmaaaz/ichor/mcp/internal/prompts"
	"github.com/timmaaaz/ichor/mcp/internal/resources"
	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func main() {
	apiURL := flag.String("api-url", "http://localhost:8080", "Ichor API base URL")
	token := flag.String("token", "", "Bearer token for Ichor API authentication")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("ICHOR_TOKEN")
	}
	if *token == "" {
		fmt.Fprintln(os.Stderr, "error: --token flag or ICHOR_TOKEN environment variable required")
		os.Exit(1)
	}

	// Create the Ichor API client.
	ichorClient := client.New(*apiURL, *token)

	// Create the MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ichor-mcp",
		Version: "0.1.0",
	}, nil)

	// Register read-only tools.
	tools.RegisterDiscoveryTools(server, ichorClient)
	tools.RegisterUIReadTools(server, ichorClient)
	tools.RegisterWorkflowReadTools(server, ichorClient)
	tools.RegisterSearchTools(server, ichorClient)

	// Register write tools.
	tools.RegisterWorkflowWriteTools(server, ichorClient)
	tools.RegisterUIWriteTools(server, ichorClient)
	tools.RegisterValidationTools(server, ichorClient)

	// Register analysis tools.
	tools.RegisterAnalysisTools(server, ichorClient)

	// Register prompts.
	prompts.RegisterPrompts(server, ichorClient)

	// Register resources.
	resources.RegisterResources(server, ichorClient)

	// Run over stdio transport.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
