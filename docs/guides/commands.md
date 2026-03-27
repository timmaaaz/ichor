# Development Commands

Run `make help` for a full list. Key commands below.

## Development Setup
```bash
make dev-gotooling    # Install Go tooling
make dev-brew         # Install Homebrew dependencies (kind, kubectl, kustomize, pgcli, watch)
make dev-docker       # Pull Docker images
```

## Testing
```bash
make test             # Run all tests with linting and vulnerability checks
make test-race        # Run tests with race detector
make test-only        # Run only tests (no linting)
make lint             # Lint code
make vuln-check       # Check for vulnerabilities
make test-down        # Shutdown test containers
```

## Local Kubernetes Development
```bash
make dev-up           # Start KIND cluster with all services
make dev-update-apply # Build containers and deploy to KIND
make dev-logs         # View logs (formatted)
make dev-logs-auth    # View auth service logs
make dev-logs-init    # View init container logs
make dev-update       # Restart deployments (after code changes)
make dev-status       # Check pod status
make dev-down         # Shutdown cluster
```

## Database Operations
```bash
make migrate          # Run migrations
make seed             # Seed database with test data
make seed-frontend    # Seed frontend configuration
make pgcli            # Access PostgreSQL CLI
make dev-database-recreate  # Recreate database (deletes all data!)
```

## Docker Compose (Alternative to Kubernetes)
```bash
make compose-up       # Start with existing images
make compose-build-up # Build and start
make compose-logs     # View logs
make compose-down     # Shutdown
```

## Running Locally (Without Containers)
```bash
make run              # Run main service locally
make run-help         # Run with help output
make admin            # Run admin tooling
```

## Authentication & API Testing
```bash
make token            # Get authentication token
export TOKEN=<TOKEN>  # Export token for subsequent requests
make users            # Test users endpoint
make curl-create      # Create new user
make live             # Test liveness probe
make ready            # Test readiness probe
make load             # Run load test (100 concurrent, 1000 requests)
```
