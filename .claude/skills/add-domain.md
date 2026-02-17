# Add Domain Skill

Guided implementation of a new domain from SQL schema. This skill parses your SQL CREATE TABLE statement and walks you through implementing all layers of the Ardan Labs architecture.

## Usage

```
/add-domain <paste your SQL schema>
```

## Your Task

When the user invokes this skill with a SQL schema, follow these steps:

### 1. Parse the SQL Schema

Extract from the CREATE TABLE statement:
- **Schema name** (e.g., `core`, `hr`, `inventory`)
- **Table name** (e.g., `pages`, `warehouses`)
- **Columns** with their types and constraints
- **Primary key** (usually `id UUID`)
- **Foreign keys** (references to other tables)
- **Unique constraints**
- **NOT NULL constraints**
- **Default values**

### 2. Derive Implementation Details

From the parsed schema, determine:

| Derived | Example |
|---------|---------|
| Entity name (singular) | `page`, `warehouse` |
| Package suffix | `pagebus`, `pageapp`, `pageapi` |
| Directory path | `business/domain/core/pagebus/` |
| DB struct name | `dbPage` (avoid conflicts with `business/sdk/page`) |
| Required fields (NOT NULL) | For `New*` struct validation |
| Optional fields | For `Update*` struct (all pointers) |

### 3. Generate Code for Each Layer

Walk through each layer in order, generating code snippets:

#### Step 1: Migration
Show the versioned SQL migration to add to `business/sdk/migrate/sql/migrate.sql`.

#### Step 2: Business Layer (`business/domain/{area}/{entity}bus/`)

Generate files:
1. **model.go** - `Entity`, `NewEntity`, `UpdateEntity` structs
2. **filter.go** - `QueryFilter` struct with pointer fields
3. **order.go** - `DefaultOrderBy` and `OrderBy*` constants
4. **event.go** - Domain events for workflow integration
5. **{entity}bus.go** - `Business` struct, `Storer` interface, CRUD methods

#### Step 3: Database Store (`business/domain/{area}/{entity}bus/stores/{entity}db/`)

Generate files:
1. **model.go** - `db{Entity}` struct with `db:` tags, conversion functions
2. **filter.go** - Filter to WHERE clause conversion
3. **order.go** - Order field mapping
4. **{entity}db.go** - `Store` struct implementing `Storer`

#### Step 4: Application Layer (`app/domain/{area}/{entity}app/`)

Generate files:
1. **model.go** - JSON structs with `json:` tags, `Encode()`/`Decode()`/`Validate()` methods
2. **filter.go** - Query parameter parsing
3. **order.go** - Order by parsing
4. **{entity}app.go** - `App` struct wrapping business layer

#### Step 5: API Layer (`api/domain/http/{area}/{entity}api/`)

Generate files:
1. **{entity}api.go** - HTTP handlers (create, update, delete, query, queryByID)
2. **route.go** - Route registration with auth middleware
3. **filter.go** - HTTP query param extraction

#### Step 6: Wiring (`api/cmd/services/ichor/build/all/all.go`)

Show:
1. Import statements to add
2. Business layer instantiation
3. Route registration

#### Step 7: Test Setup

Show:
1. Test file structure in `api/cmd/services/ichor/tests/{area}/{entity}api/`
2. Seed data example
3. Table access permissions in `business/domain/core/tableaccessbus/testutil.go`

#### Step 8: FormData Registration (Optional)

Ask if they want FormData integration. If yes, show the registration code.

### 4. Implementation Approach

For each step:
1. **Show the code** to generate
2. **Explain key decisions** (e.g., why pointers in UpdateEntity)
3. **Highlight common pitfalls** (e.g., naming conflicts with sdk packages)
4. **Wait for confirmation** before moving to the next step

### 5. Reference Documentation

Point to detailed documentation:
- [Domain Implementation Guide](docs/domain-implementation-guide.md) - Full step-by-step walkthrough
- [Layer Patterns](docs/layer-patterns.md) - Encoder/Decoder interfaces, Storer pattern
- [Debugging Guide](docs/debugging.md) - Troubleshooting common issues

## Example Session

**User**: `/add-domain CREATE TABLE hr.departments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT UNIQUE NOT NULL, code TEXT NOT NULL, manager_id UUID REFERENCES core.users(id), is_active BOOLEAN DEFAULT TRUE);`

**Claude**:
```
Parsed schema:
- Schema: hr
- Table: departments
- Columns:
  - id: UUID (PK, auto-generated)
  - name: TEXT (UNIQUE, NOT NULL)
  - code: TEXT (NOT NULL)
  - manager_id: UUID (FK → core.users)
  - is_active: BOOLEAN (default TRUE)

Implementation plan:
- Entity: Department
- Packages: departmentbus, departmentapp, departmentapi
- Directory: business/domain/hr/departmentbus/

Let's start with Step 1: Migration...
```

## Key Patterns to Apply

1. **Business layer models** use Go types (uuid.UUID, time.Time, etc.)
2. **App layer models** use strings for IDs (JSON-friendly)
3. **Update structs** use pointers for optional fields
4. **New structs** validate required fields with `validate:"required"`
5. **DB structs** use `db:` tags matching column names
6. **JSON tags** use camelCase
7. **Encoder/Decoder** interfaces for all request/response types
