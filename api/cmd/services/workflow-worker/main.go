// Package main is the entry point for the workflow-worker service.
// The workflow-worker connects to Temporal and processes workflow executions.
// Full implementation will be added in Phase 9.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var build = "develop"

func main() {
	fmt.Printf("workflow-worker starting (build: %s)\n", build)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("workflow-worker ready, waiting for shutdown signal...")
	sig := <-shutdown
	fmt.Printf("workflow-worker shutting down: signal=%s\n", sig)
}
