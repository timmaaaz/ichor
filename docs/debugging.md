# Debugging Guide

Quick reference for debugging Ichor services.

## View Database Directly

```bash
make pgcli
# Then: SELECT * FROM schema.table_name;
```

## View Logs

```bash
# Service logs (formatted)
make dev-logs

# Raw logs
kubectl logs -n ichor-system -l app=ichor --all-containers=true -f
```

## Describe Resources

```bash
make dev-describe-ichor     # Ichor pods
make dev-describe-database  # Database pod
make dev-describe-node      # Cluster nodes
```

## Common Issues

### "No keys exist"

Set `ICHOR_KEYS` environment variable or add keys to `zarf/keys/`

### Database connection fails

- Check `ICHOR_DB_HOST` matches service name
- Verify database pod is running: `make dev-status`

### Tests failing

- Ensure test database is clean: `make test-down` then `make test`
- Check migrations are current: `make migrate`

### Build fails

- Verify Go version: `go version` (must be 1.23+)
- Clean and rebuild: `go clean -modcache && go mod download && make build`

## Running a Single Test

```bash
# Run specific test function
go test -v ./api/cmd/services/ichor/tests/{area}/{entity}api -run TestFunctionName

# Run all tests in a package
go test -v ./api/cmd/services/ichor/tests/{area}/{entity}api

# Run with race detector
go test -race -v ./api/cmd/services/ichor/tests/{area}/{entity}api
```
