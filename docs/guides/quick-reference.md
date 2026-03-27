# Quick Reference

## Database Migrations

**Location**: `business/sdk/migrate/sql/migrate.sql`

**Format**:
```sql
-- Version: X.YY
-- Description: What this migration does
CREATE TABLE schema.table_name (
    id UUID PRIMARY KEY,
    ...
);
```

**Schemas** are versioned separately:
- Version 0.xx - Schema creation
- Version 1.xx - Core tables
- Version 2.xx - Configuration tables

## Configuration

**Environment Variables** (prefixed with `ICHOR_`):
- `ICHOR_DB_HOST` - Database host
- `ICHOR_DB_USER` - Database user
- `ICHOR_DB_PASSWORD` - Database password
- `ICHOR_KEYS` - RSA keys for JWT (multiline)
- `ICHOR_WEB_API_HOST` - API listen address

**Config Parsing**: Uses `github.com/ardanlabs/conf/v3`

## Observability

**Tracing**: OpenTelemetry → Tempo (Grafana stack)
- Configured in `foundation/otel/`
- 5% sampling by default

**Metrics**: Exposed via `/metrics` endpoint
- View with: `make metrics-view`
- Prometheus format

**Logging**: Structured JSON logs
- `foundation/logger/`
- Trace ID injection
- Format tool: `api/cmd/tooling/logfmt/`

**Visualization**:
- Grafana: `make grafana` (http://localhost:3100)
- Statsviz: `make statsviz` (http://localhost:3010/debug/statsviz)
