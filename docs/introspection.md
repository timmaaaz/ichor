# Database Introspection API

Read-only API for browsing the PostgreSQL schema at runtime. Used by LLM agents, the MCP server, and the frontend to discover table structure, column types, relationships, and enum types without requiring direct database access.

## Overview

The introspection API queries PostgreSQL's `information_schema` and `pg_catalog` views to expose database metadata as JSON. All endpoints are read-only and require authentication with admin role.

**Auth**: All endpoints require admin role (`auth.RuleAdminOnly`) except `queryEnumOptions` which allows any authenticated user (`auth.RuleAny`).

## Endpoints

### List Schemas

```
GET /v1/introspection/schemas
```

Returns all non-system PostgreSQL schemas.

**Response**:
```json
[
  {"name": "core"},
  {"name": "inventory"},
  {"name": "products"},
  {"name": "sales"},
  {"name": "workflow"},
  {"name": "config"}
]
```

### List Tables

```
GET /v1/introspection/schemas/{schema}/tables
```

Returns all tables in a schema with estimated row counts.

**Response**:
```json
[
  {"schema": "core", "name": "users", "row_count_estimate": 150},
  {"schema": "core", "name": "roles", "row_count_estimate": 5}
]
```

### List Columns

```
GET /v1/introspection/tables/{schema}/{table}/columns
```

Returns all columns for a table with data types, nullability, defaults, primary key status, and foreign key metadata.

**Response**:
```json
[
  {
    "name": "id",
    "data_type": "uuid",
    "is_nullable": false,
    "is_primary_key": true,
    "default_value": "gen_random_uuid()",
    "is_foreign_key": false
  },
  {
    "name": "role_id",
    "data_type": "uuid",
    "is_nullable": false,
    "is_primary_key": false,
    "default_value": "",
    "is_foreign_key": true,
    "referenced_schema": "core",
    "referenced_table": "roles",
    "referenced_column": "id"
  }
]
```

### List Relationships

```
GET /v1/introspection/tables/{schema}/{table}/relationships
```

Returns outgoing foreign key relationships (this table references others).

**Response**:
```json
[
  {
    "foreign_key_name": "fk_users_role_id",
    "column_name": "role_id",
    "referenced_schema": "core",
    "referenced_table": "roles",
    "referenced_column": "id",
    "relationship_type": "many-to-one"
  }
]
```

### List Referencing Tables

```
GET /v1/introspection/tables/{schema}/{table}/referencing-tables
```

Returns incoming foreign key relationships (other tables reference this one).

**Response**:
```json
[
  {
    "schema": "core",
    "table": "users",
    "foreign_key_column": "role_id",
    "constraint_name": "fk_users_role_id"
  }
]
```

### List Enum Types

```
GET /v1/introspection/enums/{schema}
```

Returns all PostgreSQL ENUM types in a schema with their values.

**Response**:
```json
[
  {
    "name": "order_status",
    "schema": "sales",
    "values": ["pending", "confirmed", "shipped", "delivered", "cancelled"]
  }
]
```

### Get Enum Values

```
GET /v1/introspection/enums/{schema}/{name}
```

Returns the ordered values of a specific enum type.

**Response**:
```json
[
  {"value": "pending", "sort_order": 1},
  {"value": "confirmed", "sort_order": 2},
  {"value": "shipped", "sort_order": 3}
]
```

### Get Enum Options (with Labels)

```
GET /v1/config/enums/{schema}/{name}/options
```

Returns enum values merged with human-friendly labels from `config.enum_labels`. This is the preferred endpoint for building dropdowns in the frontend.

**Auth**: Any authenticated user (not admin-only).

**Response**:
```json
[
  {"value": "pending", "label": "Pending Review", "sort_order": 1},
  {"value": "confirmed", "label": "Confirmed", "sort_order": 2},
  {"value": "shipped", "label": "Shipped", "sort_order": 3}
]
```

If no custom labels exist, `label` falls back to the raw enum value.

## Architecture

```
introspectionapi (API) → introspectionapp (App) → introspectionbus (Business) → introspectiondb (Store)
```

Follows the standard Ardan Labs layering. The store queries `information_schema.columns`, `information_schema.table_constraints`, `pg_catalog.pg_enum`, etc.

### Key Files

| File | Purpose |
|------|---------|
| `api/domain/http/introspectionapi/routes.go` | 8 route registrations |
| `api/domain/http/introspectionapi/introspectionapi.go` | 8 handlers |
| `app/domain/introspectionapp/model.go` | Response types (Schema, Table, Column, Relationship, EnumType, EnumValue, EnumOption) |
| `app/domain/introspectionapp/introspectionapp.go` | App layer (conversion + delegation) |
| `business/domain/introspectionbus/introspectionbus.go` | Business logic |
| `business/domain/introspectionbus/stores/introspectiondb/` | SQL queries against information_schema |

### Database Schemas

The introspection API covers these PostgreSQL schemas:

| Schema | Domain |
|--------|--------|
| `core` | Users, roles, permissions, contact info |
| `hr` | Offices, titles, reports-to |
| `geography` | Countries, regions, cities, streets |
| `assets` | Asset types, conditions, user assets |
| `inventory` | Warehouses, zones, locations, items |
| `products` | Products, brands, categories, costs |
| `procurement` | Suppliers, supplier products |
| `sales` | Customers, orders, line items |
| `config` | Table configs, page configs, forms |
| `workflow` | Automation rules, actions, edges |

## Usage Patterns

### Agent: Building a Form

1. `GET /v1/introspection/schemas` → pick target schema
2. `GET /v1/introspection/schemas/sales/tables` → find `orders` table
3. `GET /v1/introspection/tables/sales/orders/columns` → get column types
4. Map columns to form field types (uuid→hidden, text→text, numeric→number, etc.)
5. `GET /v1/introspection/tables/sales/orders/relationships` → identify FK dropdowns

### Agent: Building a Table Config

1. `GET /v1/introspection/tables/sales/orders/columns` → column inventory
2. `GET /v1/introspection/tables/sales/orders/relationships` → identify join targets
3. Use column data_type to set `visual_settings.columns[].type` (text→string, numeric→number, etc.)

### MCP: search_database_schema Tool

The MCP `search_database_schema` tool wraps these endpoints progressively:
- No args → calls `GET /v1/introspection/schemas`
- Schema only → calls `GET /v1/introspection/schemas/{schema}/tables`
- Schema + table → calls both columns and relationships, merges into single response
